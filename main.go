package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

var pstates = []uint64{0xC0010064, 0xC0010065, 0xC0010066, 0xC0010067, 0xC0010068, 0xC0010069, 0xC001006A, 0xC001006B}

func main() {
	list := flag.Bool("list", false, "List all P-States")
	pstate := flag.Int("pstate", -1, "P-State to set")
	enable := flag.Bool("enable", false, "Enable P-State")
	disable := flag.Bool("disable", false, "Disable P-State")
	fid := flag.String("fid", "-1", "FID to set (in hex)")
	did := flag.String("did", "-1", "DID to set (in hex)")
	vid := flag.String("vid", "-1", "VID to set (in hex)")
	c6Enable := flag.Bool("c6-enable", false, "Enable C-State C6")
	c6Disable := flag.Bool("c6-disable", false, "Disable C-State C6")

	flag.Parse()

	if *list {
		for i, ps := range pstates {
			val, err := readMSR(ps, 0)
			if err != nil {
				fmt.Printf("Error reading P-State %d: %v\n", i, err)
				continue
			}
			fmt.Printf("P%d - %s\n", i, pstateToStr(val))
		}
		val, err := readMSR(0xC0010292, 0)
		if err != nil {
			fmt.Printf("Error reading C6 state: %v\n", err)
		} else {
			c6Package := "Disabled"
			if val&(1<<32) != 0 {
				c6Package = "Enabled"
			}
			fmt.Println("C6 State - Package - " + c6Package)

			val, err = readMSR(0xC0010296, 0)
			if err != nil {
				fmt.Printf("Error reading C6 core state: %v\n", err)
			} else {
				c6Core := "Disabled"
				if val&((1<<22)|(1<<14)|(1<<6)) == ((1 << 22) | (1 << 14) | (1 << 6)) {
					c6Core = "Enabled"
				}
				fmt.Println("C6 State - Core - " + c6Core)
			}
		}
	}

	if *pstate >= 0 {
		oldVal, err := readMSR(pstates[*pstate], 0)
		if err != nil {
			fmt.Println("Error reading MSR:", err)
			return
		}
		fmt.Printf("Current P%d: %s\n", *pstate, pstateToStr(oldVal))

		newVal := oldVal
		if *enable {
			newVal = setBits(newVal, 63, 1, 1)
			fmt.Println("Enabling state")
		}
		if *disable {
			newVal = setBits(newVal, 63, 1, 0)
			fmt.Println("Disabling state")
		}
		if *fid != "-1" {
			fidVal, _ := strconv.ParseUint(*fid, 16, 64)
			newVal = setFID(newVal, fidVal)
			fmt.Printf("Setting FID to %X\n", fidVal)
		}
		if *did != "-1" {
			didVal, _ := strconv.ParseUint(*did, 16, 64)
			newVal = setDID(newVal, didVal)
			fmt.Printf("Setting DID to %X\n", didVal)
		}
		if *vid != "-1" {
			vidVal, _ := strconv.ParseUint(*vid, 16, 64)
			newVal = setVID(newVal, vidVal)
			fmt.Printf("Setting VID to %X\n", vidVal)
		}
		if newVal != oldVal {
			err := writeMSR(pstates[*pstate], newVal, -1)
			if err != nil {
				fmt.Println("Error writing MSR:", err)
				return
			}
			fmt.Printf("New P%d: %s\n", *pstate, pstateToStr(newVal))
		}
	}

	if *c6Enable {
		enableC6State(true)
	}

	if *c6Disable {
		enableC6State(false)
	}
}

func writeMSR(msr uint64, val uint64, cpu int) error {
	var path string
	if cpu == -1 {
		cpus, err := filepath.Glob("/dev/cpu/*/msr")
		if err != nil {
			return err
		}
		for _, cpuPath := range cpus {
			if err := writeMSRToFile(cpuPath, msr, val); err != nil {
				fmt.Printf("Warning: Failed to write MSR for path %s: %v\n", cpuPath, err)
				// Continue attempting to write to other CPUs despite the error
			}
		}
		return nil // Return nil assuming at least some writes were successful
	} else {
		path = fmt.Sprintf("/dev/cpu/%d/msr", cpu)
		return writeMSRToFile(path, msr, val)
	}
}

func writeMSRToFile(path string, msr uint64, val uint64) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	// Correct the way to convert uint64 to a byte slice.
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, val)
	_, err = file.WriteAt(data, int64(msr))
	return err
}

func readMSR(msr uint64, cpu int) (uint64, error) {
	path := fmt.Sprintf("/dev/cpu/%d/msr", cpu)
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	data := make([]byte, 8)
	_, err = file.ReadAt(data, int64(msr))
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(data), nil
}

func pstateToStr(val uint64) string {
	if val&(1<<63) != 0 {
		fid := val & 0xff
		did := (val & 0x3f00) >> 8
		vid := (val & 0x3fc000) >> 14
		ratio := 25 * float64(fid) / (12.5 * float64(did))
		vcore := 1.55 - 0.00625*float64(vid)
		return fmt.Sprintf("Enabled - FID = %X - DID = %X - VID = %X - Ratio = %.2f - vCore = %.5f", fid, did, vid, ratio, vcore)
	} else {
		return "Disabled"
	}
}

func setBits(val uint64, base, length, new uint64) uint64 {
	// Ensure all operands in the bitwise operations are of the same type.
	mask := (uint64(1)<<length - 1) << base
	return (val &^ mask) | (new << base)
}

func setFID(val, new uint64) uint64 {
	return setBits(val, 0, 8, new)
}

func setDID(val, new uint64) uint64 {
	return setBits(val, 8, 6, new)
}

func setVID(val, new uint64) uint64 {
	return setBits(val, 14, 8, new)
}

func enableC6State(enable bool) {
	c6RegPackage := uint64(0xC0010292)
	c6RegCore := uint64(0xC0010296)
	enableBit := uint64(1 << 32)
	coreEnableBits := uint64((1 << 22) | (1 << 14) | (1 << 6))

	val, err := readMSR(c6RegPackage, 0)
	if err != nil {
		fmt.Printf("Error reading C6 package state: %v\n", err)
		return
	}

	if enable {
		val |= enableBit
	} else {
		val &^= enableBit
	}
	writeMSR(c6RegPackage, val, -1)

	val, err = readMSR(c6RegCore, 0)
	if err != nil {
		fmt.Printf("Error reading C6 core state: %v\n", err)
		return
	}

	if enable {
		val |= coreEnableBits
	} else {
		val &^= coreEnableBits
	}
	writeMSR(c6RegCore, val, -1)

	if enable {
		fmt.Println("Enabled C6 state")
	} else {
		fmt.Println("Disabled C6 state")
	}
}
