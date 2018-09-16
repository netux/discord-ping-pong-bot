package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// CommandExecutionError represents an error raised
// on command execution due to the user or input being wrong.
type CommandExecutionError struct {
	s string
}

func (err *CommandExecutionError) Error() string {
	return err.s
}

var (
	// ErrWrongTurn represents the error sent when an user attempts to hit
	// the ball when it's not their turn.
	ErrWrongTurn = CommandExecutionError{"it's not your turn"}
	// ErrNoAvailableGames represents the sent when an user attempts to pong
	// but they are no games to join.
	ErrNoAvailableGames = CommandExecutionError{"no ping-pong games to join"}
)

func handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer func() {
		if r := recover(); r != nil {
			var errMsg string
			if err, ok := r.(error); ok {
				errMsg = "internal error: " + err.Error()
			} else if err, ok := r.(CommandExecutionError); ok {
				errMsg = err.Error()
			} else {
				errMsg = fmt.Sprintf("%v", r)
			}

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ %s %s", m.Author.Mention(), errMsg))
		}
	}()

	if s.State == nil {
		return
	}

	if m.GuildID == "" {
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(Cfg.ChannelWhitelist) > 0 {
		inWhitelist := func() bool {
			for _, chID := range Cfg.ChannelWhitelist {
				if chID == m.ChannelID {
					return true
				}
			}

			return false
		}()

		if !inWhitelist {
			return
		}
	}

	reMatch := RegexpPingpong.FindStringSubmatch(m.Content)
	if reMatch == nil {
		return
	}

	pingPongStr := reMatch[1]
	msg := reMatch[2]

	ppMatchIndex, ppMatch := GetPingpongMatchWithUser(m.Author.ID)

	// Stop a match when one player sends "🏓 exit"
	if ppMatchIndex >= 0 && strings.ToLower(msg) == "exit" {
		user1, err := s.User(ppMatch.PlayerIDs[0])
		if err != nil {
			panic(err)
		}

		mentions := fmt.Sprintf("%s", user1.Mention())

		if ppMatch.PlayerIDs[1] != "" {
			user2, err := s.User(ppMatch.PlayerIDs[1])
			if err != nil {
				panic(err)
			}

			mentions += fmt.Sprintf(" %s", user2.Mention())
		}

		PingpongMatches = append(PingpongMatches[:ppMatchIndex], PingpongMatches[ppMatchIndex+1:]...)
		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("ℹ %s match ended", mentions))
		if err != nil {
			panic(err)
		}
		return
	}

	if isPing(pingPongStr) {
		if ppMatch == nil {
			// Create the match.
			newPPMatch := NewPingpongMatch(m.Author.ID, m.ChannelID)
			ppMatch = &newPPMatch
			PingpongMatches = append(PingpongMatches, ppMatch)

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("ℹ %s started a game, waiting for player 2 to respond with %s", m.Author.Mention(), Cfg.PongPrefix))
			return
		}

		if !ppMatch.Started {
			panic(CommandExecutionError{"can't ping yet, waiting for player 2"})
		}

		if ppMatch.LastHitPing {
			panic(ErrWrongTurn)
		}
	} else if isPong(pingPongStr) {
		if ppMatch == nil {
			if len(PingpongMatches) > 0 {
				// Join the match.
				ppMatch = GetNextAvailablePingpongMatch()
				if ppMatch == nil {
					panic(ErrNoAvailableGames)
				} else {
					ppMatch.SetSecondUser(m.Author.ID)

					user1, err := s.User(ppMatch.PlayerIDs[0])
					if err != nil {
						panic(err)
					}

					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("ℹ %s joined the match. %s serves.", m.Author.Mention(), user1.Mention()))
					return
				}
			} else {
				panic(ErrNoAvailableGames)
			}
		}

		if ppMatch.PlayerIDs[1] == "" && ppMatch.PlayerIDs[0] == m.Author.ID {
			panic(CommandExecutionError{"can't play by yourself, sorry"})
		}

		if !ppMatch.LastHitPing {
			panic(ErrWrongTurn)
		}
	}

	hit1, hit2, hitOk := CalculatePingpongTableHit(msg)
	ppMatch.Hit(hitOk)

	var response string

	if hitOk {
		response = fmt.Sprintf("✔ Ball hit %s! At %.2f%% and %.2f%%", m.Author.Mention(), hit1, hit2)
	} else {
		if hit2 == 0 {
			response = fmt.Sprintf("✖ Ball hit too far (at %.2f%%) %s.", hit1, m.Author.Mention())
		} else {
			response = fmt.Sprintf("✖ Ball hit once at %.2f%% but then it bounced too far (at %.2f%%) %s.", hit1, hit2, m.Author.Mention())
		}

		response += "\n" + fmt.Sprintf("**Score:** %d - %d", ppMatch.Scores[0], ppMatch.Scores[1])
	}

	_, err := s.ChannelMessageSend(m.ChannelID, response)
	if err != nil {
		panic(err)
	}
}
