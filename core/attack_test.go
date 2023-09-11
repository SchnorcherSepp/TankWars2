package core

import (
	"fmt"
	"testing"
	"time"
)

func Test_print_calcDamage(t *testing.T) {

	for armour := 0; armour <= 4; armour++ {
		fmt.Printf("\n")
		for _, demoralized := range []bool{false, true} {

			mn := 999999
			mx := 0
			sum := 0
			avr := 0.0
			crt := 0
			cp := 0.0

			for i := 1; i <= 1000000; i++ {
				damage, c := calcDamage(demoralized, armour)
				if mn > damage {
					mn = damage
				}
				if mx < damage {
					mx = damage
				}
				sum += damage
				avr = float64(sum) / float64(i)
				if c {
					crt++
				}
				cp = float64(crt) / float64(i) * 100
			}

			fmt.Printf("Armour: %d [%v]\t\t%d < %.0f < %d\t[%.0f%%]\n", armour, demoralized, mn, avr, mx, cp)
			time.Sleep(1 * time.Second)
		}
	}
}
