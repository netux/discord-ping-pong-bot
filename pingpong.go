package main

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

// RegexpPingpong matches "🏓", "🏓 {message}" and it's variants.
var RegexpPingpong *regexp.Regexp

// SetRegexpPingpong compiles a new Regexp and stores it in RegexpPingpong.
func SetRegexpPingpong(pingPrefix, pongPrefix string) {
	RegexpPingpong = regexp.MustCompile(
		fmt.Sprintf(
			"^(%s|%s)(?: (.+))?$",
			regexp.QuoteMeta(pingPrefix),
			regexp.QuoteMeta(pongPrefix),
		),
	)
}

func isPing(s string) bool {
	return s == Cfg.PingPrefix
}

func isPong(s string) bool {
	return s == Cfg.PongPrefix
}

// PingpongMatches stores all active ping pong matches.
var PingpongMatches = []*PingpongMatch{}

// PingpongMatch stores a ping pong match's information.
type PingpongMatch struct {
	PlayerIDs              [2]string
	Scores                 [2]uint
	ChannelID              string
	Started, LastHitPing   bool
	StartTime, LastHitTime time.Time
}

// NewPingpongMatch creates a new PingpongMatch.
func NewPingpongMatch(startUserID, channelID string) PingpongMatch {
	return PingpongMatch{
		PlayerIDs:   [2]string{startUserID},
		Scores:      [2]uint{},
		ChannelID:   channelID,
		LastHitPing: false,
		StartTime:   time.Time{},
		LastHitTime: time.Time{},
	}
}

func (m *PingpongMatch) start() {
	m.Started = true
	m.StartTime = time.Now()
}

// SetSecondUser sets the second user's ID in PingpongMatch.
// It panics if the second user is already set or if it's equal to the first user.
func (m *PingpongMatch) SetSecondUser(userID string) {
	if m.PlayerIDs[0] == userID {
		panic(errors.New("second user can't be the same as first user"))
	} else if m.PlayerIDs[1] != "" {
		panic(errors.New("second user already set"))
	} else {
		m.PlayerIDs[1] = userID
	}

	m.start()
}

// Hit updates LastHitTime and increases scores.
func (m *PingpongMatch) Hit(ok bool) {
	m.LastHitTime = time.Now()
	if !ok {
		if m.LastHitPing {
			m.Scores[0]++
		} else {
			m.Scores[1]++
		}
	}
	m.LastHitPing = !m.LastHitPing
}

// GetPingpongMatchWithUser returns the active PingpongMatch where one of it's
// users is userID, nil otherwise.
func GetPingpongMatchWithUser(userID string) (int, *PingpongMatch) {
	for i, m := range PingpongMatches {
		if m.PlayerIDs[0] == userID || m.PlayerIDs[1] == userID {
			return i, m
		}
	}

	return -1, nil
}

// GetNextAvailablePingpongMatch goes through PingpongMatches backwards and
// returns the first match where the second user is not set.
func GetNextAvailablePingpongMatch() *PingpongMatch {
	for i := range PingpongMatches {
		m := PingpongMatches[len(PingpongMatches)-1-i]

		if m.PlayerIDs[1] == "" {
			return m
		}
	}

	return nil
}

const (
	// HitEasiness is the factor used to multiply and divide the Z coordinate of the
	// pingpong hit to make it easier to hit right
	HitEasiness = 3
)

// CalculatePingpongTableHit calculates whether a hit was successful based on the hash
// of msg, which is used as the seed for a RNG. The RNG then calculates a percentage which
// represents the Z coordinate at which a real ball would hit the table.
func CalculatePingpongTableHit(msg string) (float32, float32, bool) {
	src := rand.NewSource(Hash(msg) + rand.Int63())
	rng := rand.New(src)

	hit1 := float32(rng.Intn(100*HitEasiness) + 1)
	actualHit1 := hit1 / HitEasiness
	if hit1 >= 75*HitEasiness {
		return actualHit1, 0, false
	}

	hit2 := float32(rng.Intn(75*HitEasiness) + 1)
	actualHit2 := hit2 / HitEasiness
	return actualHit1, 50 + actualHit2, hit2 < 50*HitEasiness
}
