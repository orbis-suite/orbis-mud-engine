package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/mud/config"
	"example.com/mud/parser/commands"
	orbisplugin "example.com/mud/plugin"
	"example.com/mud/server"
	"example.com/mud/world"
	"example.com/mud/world/entities"
	"example.com/mud/world/player"
)

func handleConnection(conn net.Conn, gameWorld *world.World, cfg *config.Config) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	var name string

	for {
		if _, err := fmt.Fprint(conn, "What is your name, weary adventurer? "); err != nil {
			return
		}

		name, _ = reader.ReadString('\n')
		name = strings.TrimSpace(name)

		vdn := player.NameValidation(name)
		if vdn != "" {
			fmt.Fprint(conn, vdn)
			continue
		}
		break
	}

	inbox := make(chan string, 64)
	p, err := gameWorld.AddPlayer(name, inbox)
	if err != nil {
		err := fmt.Errorf("error adding player: %w", err)
		fmt.Println(err.Error())
		fmt.Fprintln(conn, err.Error())
		return
	}

	opening, err := p.OpeningMessage()
	if err != nil {
		err := fmt.Errorf("error printing opening message: %w", err)
		fmt.Println(err.Error())
		fmt.Fprintln(conn, err.Error())
		return
	}
	rendered, err := player.RenderForTelnet(opening)
	if err != nil {
		fmt.Fprintln(conn, err.Error())
		return
	}
	fmt.Fprintln(conn, rendered)

	// start consuming incoming messages
	go handleConnectionIncoming(conn, inbox)

	// notify when outgoing messages end
	done := make(chan struct{})
	go func() {
		handleConnectionOutgoing(conn, gameWorld, p, cfg)
		close(done)
	}()

	// Wait until the outgoing loop ends
	<-done
}

func handleConnectionIncoming(conn net.Conn, inbox chan string) {
	go func() {
		for msg := range inbox {
			// Use CRLF for telnet clients
			fmt.Fprint(conn, msg+"\r\n")
		}
	}()
}

func handleConnectionOutgoing(conn net.Conn, gameWorld *world.World, p *player.Player, cfg *config.Config) {
	scanner := bufio.NewScanner(conn)
	for {
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.ToLower(line) == "quit" {
			break
		}

		// check if player has pending multi-part messages
		if pending := p.Pending; pending != nil {
			n, nerr := strconv.Atoi(line)
			if nerr == nil {
				slot := pending.Ambiguity.Slots[pending.StepIndex]
				if n >= 1 && n <= len(slot.Matches) {
					pending.Selected[slot.Role] = n - 1
					pending.StepIndex++
					if pending.StepIndex < len(pending.Ambiguity.Slots) {
						// continue prompting current slot
						promptCurrentSlot(conn, pending)
					} else {
						chosen := make(map[string]*entities.Entity, len(pending.Selected))
						for _, s := range pending.Ambiguity.Slots {
							idx := pending.Selected[s.Role]
							chosen[s.Role] = s.Matches[idx].Entity
						}
						out, execErr := pending.Ambiguity.Execute(chosen)
						p.Pending = nil
						if execErr != nil {
							fmt.Fprintln(conn, execErr.Error())
						} else if out != "" {
							fmt.Fprintln(conn, out)
						}
					}
				}
				continue
			}
			// non-number input falls through and removes pending action
			p.Pending = nil
		}

		if coolDownTime := p.CooldownRemaining(); coolDownTime > 0 {
			fmt.Fprintf(conn, "You need to catch your breath. Try again in %.1fs\r\n", coolDownTime.Seconds())
			continue
		}

		resp, err := gameWorld.Parse(p, line)
		if err != nil {
			var amb *entities.AmbiguityError
			if errors.As(err, &amb) {
				p.Pending = &entities.PendingAction{
					Ambiguity: amb,
					StepIndex: 0,
					Selected:  map[string]int{},
				}
				promptCurrentSlot(conn, p.Pending)
				continue
			}
			err := fmt.Errorf("error received: %w", err)
			fmt.Println(err.Error())
			fmt.Fprintln(conn, err.Error())
		} else if resp != nil {
			rendered, renderErr := player.RenderForTelnet(resp)
			if renderErr != nil {
				fmt.Fprintln(conn, renderErr.Error())
			} else if rendered != "" {
				fmt.Fprintln(conn, rendered)
			}
		}

		p.StartCooldown(time.Duration(cfg.PlayerRateLimit) * time.Millisecond)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Connection error:", err)
	}

	gameWorld.DisconnectPlayer(p)
	fmt.Printf("Connection closed\n")
}

func promptCurrentSlot(conn net.Conn, p *entities.PendingAction) {
	slot := p.Ambiguity.Slots[p.StepIndex]
	fmt.Fprintln(conn, slot.Prompt)
	for i, opt := range slot.Matches {
		fmt.Fprintf(conn, "  %d) %s\r\n", i+1, opt.Text)
	}
}

func main() {
	debug := flag.Bool("debug", false, "connect to a game binary already running with -debug instead of launching a subprocess")
	flag.Parse()

	// load configuration file
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var gameClient orbisplugin.GameClient
	var cleanup func()
	if *debug {
		gameClient, cleanup, err = orbisplugin.ConnectDebug(5 * time.Second)
	} else {
		gameClient, cleanup, err = orbisplugin.Launch(cfg.GameBinary)
	}
	if err != nil {
		log.Fatalf("failed to connect to game: %v", err)
	}
	defer cleanup()

	manifest, err := gameClient.GetManifest(context.Background())
	if err != nil {
		log.Fatalf("failed to get game manifest: %v", err)
	}

	entityMap, cmds, err := orbisplugin.ManifestToWorld(manifest, gameClient)
	if err != nil {
		log.Fatalf("failed to build world from manifest: %v", err)
	}

	// validate starting room exists in entity map
	if _, ok := entityMap[manifest.GetStartingRoom()]; !ok {
		log.Fatalf("room '%s' does not exist in world.", manifest.GetStartingRoom())
	}

	if err := commands.RegisterBuiltInCommands(); err != nil {
		log.Fatalf("failed to register built-in commands: %v", err)
	}

	if err := commands.RegisterCommands(cmds); err != nil {
		log.Fatalf("failed to register DSL commands: %v", err)
	}

	gameWorld := world.NewWorld(entityMap, manifest.GetStartingRoom())

	go func() {
		addr := fmt.Sprintf(":%d", cfg.WebSocketPort)
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			server.HandleWS(w, r, gameWorld, cfg)
		})
		fmt.Printf("WebSocket server listening on port %d...\n", cfg.WebSocketPort)
		log.Fatal(http.ListenAndServe(addr, nil))
	}()

	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("MUD server listening on port 4000...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, gameWorld, cfg)
	}
}
