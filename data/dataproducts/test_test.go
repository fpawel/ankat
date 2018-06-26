package dataproducts

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/fpawel/goanalit"
	"path/filepath"
	"sync"
	"time"
	"fmt"
)

func TestNewParty(t *testing.T) {
	filename := filepath.Join( goanalit.appFolder("tests").Path(), "products.db" )

	db := MustOpen(filename, []string{"035", "035-1", "100-1"})

	p0 := db.CurrentParty()

	assert.True(t, db.PartyExists(p0.PartyID),   "current party should exist")
	assert.Equal(t, p0, db.Party(p0.PartyID),  "current parties should be equal")
	assert.Equal(t, p0, db.CurrentParty(),  "current party should be equal")

	p1 := db.NewParty( NewPartyConfig{4, "035", KeysValues{"gas1":123.456}})

	assert.True(t, db.PartyExists(p1.PartyID),   "new party should exist")

	p2 := db.CurrentParty()
	assert.Equal(t, p1, p2,  "new party should be current")
	assert.Equal(t, len(p1.Products), len(p2.Products),  "lengths should be equal")
	for i := range p1.Products{
		assert.Equal(t, p1.Products[i], p2.Products[i],  "products should be equal")
	}

	db.SetPartyValue("gas2", 789.012)

	var v float64
	err := db.GetPartyValue(p1.PartyID, "gas1", &v)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 123.456, v,   "param value should be equal")

	err = db.GetPartyValue(p1.PartyID, "gas2", &v)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 789.012, v,   "param value should be equal")


	db.DeleteParty( p2.PartyID)

	assert.True(t, !db.PartyExists(p2.PartyID),   "deleted party should not exist")

	assert.Equal(t, p0, db.CurrentParty(),  "parties should be equal")
}

func TestConcurrentRead(t *testing.T) {
	filename := filepath.Join( goanalit.appFolder("tests").Path(), "products.db" )

	db := MustOpen(filename, []string{"035", "035-1", "100-1"})

	party0 := db.CurrentParty()
	ch := make (chan Party)
	for i := 0; i <1000; i++{
		go func () {
			ch <- db.CurrentParty()
		} ()
	}

	n := 0
	for p := range ch {
		assert.Equal(t, party0, p,  "parties should be equal")
		if n++; n == 1000{
			break
		}
	}
}

func TestConcurrentWrite(t *testing.T) {
	filename := filepath.Join( goanalit.appFolder("tests").Path(), "products.db" )
	db := MustOpen(filename, []string{"035", "035-1", "100-1"})
	var mu sync.Mutex
	done := make (chan time.Duration)
	const iterationCount = 100
	for i := 0; i <iterationCount; i++{
		go func() {

			mu.Lock()
			start := time.Now()
			p := db.NewParty(NewPartyConfig{4, "035", KeysValues{"gas1":123.456}})

			assert.True(t, db.PartyExists(p.PartyID),   "party should exist")
			assert.Equal(t, p, db.Party(p.PartyID),  "parties should be equal")
			assert.Equal(t, p, db.CurrentParty(),  "current party should be equal")

			db.DeleteParty( p.PartyID)

			assert.True(t, !db.PartyExists(p.PartyID),   "party should not exist")

			done <- time.Since(start)

			mu.Unlock()
		}()
	}

	i := 0
	for t := range done{
		i++
		fmt.Println("[", i, "]", t)
		if i == iterationCount {
			break
		}
	}
}