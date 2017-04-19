package main

import(
	"fmt"
	"os"
)

const WORDBITS uint = 32

func countBitDifferences(a, b int) int {
    var sum int = 0
    var d int = a ^ b

    for d > 0 {
        sum += (d & 1)
        d >>= 1
    }

    return sum
}

/************************************************************
 * leftRotate(i int, n uint) int
 *
 * Rotates an integer, i, by n bits to the left wrapping 
 * around 
 ***********************************************************/
func leftRotate(i int, n uint) int {
	return ((i << n) | (i >> (WORDBITS - n))) & 0xffffffff
}

/************************************************************
 * expand(b []int) []int 
 *
 * Expands a slice of 16 ints to a slice of 80 ints by using 
 * the sha-1 expansion alg.
 ***********************************************************/
func expand(b []int) []int {
	res := make([]int , len(b))
    copy(res, b)
	for i := 16; i < 80; i++ {
		res = append(res, leftRotate(res[i-3] ^ res[i-8] ^ res[i-14] ^
		 res[i-16], 1))
	}

	return res
}

/************************************************************
 * mod (int, int) int
 *
 * returns a %n as it should be, instead of remainder
 ************************************************************/
func mod(a, n int) int {
	return (((a % n) + n) % n)
}

/************************************************************
 * pad (s []byte) []byte 
 *
 * pads a slice of bytes by appending the byte 0x80 and then
 * the byte 0x00 until the length of our appended slice is 
 * equivalent to 56 modulo 64 (leaving 8 bytes free). 
 * The length of the original byte string is then appended 
 * as 8 bytes (in big-endian)
 ***********************************************************/
func pad(s []byte) []byte{
	var res []byte = append(s, 0x80)
	var byteLen uint64 = uint64(uint64(len(s)) * 8)

	for i := 0; i < mod((56 - (len(s)+1)), 64); i++ {
		res = append(res, 0x00)
	}

	for i := 0; i < 8; i++ {
		var shift uint = uint((7 - i) * 8)
		// fmt.Printf("i = %d", i)
		// fmt.Printf("\tstr = 0x%02x", byteLen)
		// fmt.Printf("\tsft = 0x%02x\n", uint64(0xff) << shift)
		// fmt.Printf("\t0x%02x\n", (byteLen & (uint64(0xff) << shift)) >> shift)
		res = append(res, byte((byteLen & (uint64(0xff) << shift)) >> shift))
	}

	return res
}

/************************************************************
 * wordify(s []byte) []int 
 * 
 * Converts a slice of bytes into a slice of words. Reads 
 * four bytes at a time and shifts them appropriately to form
 * a word. Appends the collection of words. 
 * Does a sanitation check to ensure input is a mulitple of 
 * 4 bytes.
 ***********************************************************/
func wordify(s []byte) []int {
	var ret []int

	if (len(s) % 4 != 0){
		return ret
	}

	for i := 0; i <= len(s)-4; i += 4 {
		var val int = (int(s[i]) << 24) ^ (int(s[i+1]) << 16) ^ 
			(int(s[i+2]) << 8) ^ int(s[i+3])
		ret = append(ret, val)
	}

	return ret
}

/***********************************************************
 * chunkify(s []byte) [][]int
 *
 * Splits a slice of bytes into a double slice of ints. 
 * First converts the byte slice into an int slice using 
 * wordify. Then forms groups of 16 words (512 bits) and 
 * returns those.
 **********************************************************/
func chunkify(s []byte) [][] int {
	var words []int = wordify(s)
	var spl [][] int

	for i := 0; i < len(words)/16; i++ {
		spl = append(spl, words[16*i:16*i+16])
	}

	return spl
}

/**********************************************************
 * getVals(round int, b int, c int, d int) int, int 
 *
 * computes the non-linear functions necessary based on 
 * round and b,c,d values
 *********************************************************/
func getVals(round int, b int, c int, d int) (int, int) {
    var f, k int

    if (round < 20){
        f = d ^ (b & (c ^ d))
        k = 0x5a827999
    }else if (round < 40){
        f = b ^ c ^ d
        k = 0x6ed9eba1
    }else if (round < 60){
        f = (b & c) | (b & d) | (c & d)
        k = 0x8f1bbcdc
    }else{
        f = b ^ c ^ d
        k = 0xca62c1d6
    }

    return f, k
}

/**********************************************************
 * round(itr int, currState [5] int, word int) [5]int
 *
 * Computes a given round of sha-1. Gets the needed values
 * and then shifts (and rotates) values.
 *********************************************************/
func round(itr int, currState [5] int, word int) [5]int{
    f, k := getVals(itr, currState[1], currState[2], currState[3])
    var ret [5]int
    ret[0] = ((leftRotate(currState[0], 5)) + f + currState[4] + k + word) &
        0xffffffff
    ret[1] = currState[0]
    ret[2] = leftRotate(currState[1], 30)
    ret[3] = currState[2]
    ret[4] = currState[3]
    // fmt.Printf("round %d\n", itr)
    // fmt.Printf("\tf = 0x%08x, k = 0x%08x\n", f, k)
    // fmt.Printf("\t(leftRotate(currState[0], 5)) = 0x%08x\n", 
    //     leftRotate(currState[0], 5))
    // fmt.Printf("\tcurrState[4] = 0x%08x\n", currState[4])
    // fmt.Printf("\tword = 0x%08x\n", word)

    return ret 
}

/***********************************************************
 * test(f, s [5]int, r int) bool
 *
 * test whether f and s agree on the first component, if so
 * print the round and the whole f and s
 **********************************************************/
func test(f, s [5]int, r int) bool {
    if (f[0] == s[0]){
        fmt.Printf("local Collision found, round %2d:\n", r)
        for i := 0; i < len(f); i++ {
            fmt.Printf("\tf[%d] = 0x%08x\n", i, f[i])
            fmt.Printf("\ts[%d] = 0x%08x\n", i, s[i])
        }

        return true
    }

    return false
}

func main(){
	var (
		msg []byte
		args []string
		chunks [][]int
        tState [5]int
        tpState [5]int
	)

	args = os.Args[1:]

	if (len(args) == 0){
		msg = []byte("hi")
	} else{
		msg = []byte(args[0] + "\n")
	}

	// fmt.Println("len of mesage (before padding) =", len(msg))
	msg = pad(msg)
	// fmt.Println("len of mesage (after padding)  =", len(msg))
	chunks = chunkify(msg)

	state := [5]int{0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476, 0xc3d2e1f0}
    pState := [5]int{0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476, 0xc3d2e1f0}
	for _, chunk := range chunks {
        pChunk := make([]int, len(chunk))
        copy(pChunk, chunk)
        // do wang's disturbance vector stuff
        pChunk[0] = pChunk[0] ^ (0x40000001)
        pChunk[1] = pChunk[1] ^ (0x2)
        pChunk[2] = pChunk[2] ^ (0x2)
        pChunk[3] = pChunk[3] ^ (0x80000002)
        pChunk[4] = pChunk[4] ^ (0x1)

        pChunk[6] = pChunk[6] ^ (0x80000001)
        pChunk[7] = pChunk[7] ^ (0x2)
        pChunk[8] = pChunk[8] ^ (0x2)
        pChunk[9] = pChunk[9] ^ (0x2)


        pChunk[12] = pChunk[12] ^ (0x1)

        pChunk[14] = pChunk[14] ^ (0x80000002)
        pChunk[15] = pChunk[15] ^ (0x2)
		// fmt.Println(state)
        chunk = []int {0x132b5ab6, 0xa115775f, 0x5bfddd6b, 0x4dc470eb, 0x0637938a,
            0x6cceb733, 0x0c86a386, 0x68080139, 0x534047a4, 0xa42fc29a, 
            0x06085121, 0xa3131f73, 0xad5da5cf, 0x13375402, 0x40bdc7c2,
            0xd5a839e2}
        pChunk = []int{0x332b5ab6, 0xc115776d, 0x3bfddd28, 0x6dc470ab, 0xe63793c8,
            0x0cceb731, 0x8c86a387, 0x68080119, 0x534047a7, 0xe42fc2c8, 
            0x46085161, 0x43131f21, 0x0d5da5cf, 0x93375442, 0x60bdc7c3,
            0xf5a83982}
        // chunk = []int {0xb65a2b13, 0x5f7715a1, 0x6bddfd5b, 0xeb70c44d, 0x8a933706,
        //     0x33b7ce6c, 0x86a3860c, 0x39010868, 0xa4474053, 0x9ac22fa4, 
        //     0x21510806, 0x731f13a3, 0xcfa55dad, 0x02543713, 0xc2c7bd40,
        //     0xe239a6d5}
        // pChunk = []int{0x332b5ab6, 0xc115776d, 0x3bfddd28, 0x6dc470ab, 0xe63793c8,
        //     0x0cceb731, 0x8c86a387, 0x68080119, 0x534047a7, 0xe42fc2c8, 
        //     0x46085161, 0x43131f21, 0x0d5da5cf, 0x93375442, 0x60bdc7c3,
        //     0xf5a83982}
		var w []int = expand(chunk)
        var wp []int = expand(pChunk)

        // for verifying that the modified wp matches the disturbance vector
        // (it does)

        // for i := 0; i < len(w); i++ {
        //     fmt.Printf("w[%2d] ^ wp[%2d] = 0x%x\n", i, i, w[i] ^ wp[i])
        // }

        // wp := make([]int, len(w))
        // copy(wp, w)

        // flip the 1 bit (second least significant bit) in 1st word 
        // wp[0] = wp[0] ^ (1 << 1)
        // wp[1] = wp[1] ^ (1 << 6)
        // wp[2] = wp[2] ^ (1 << 1)
        // wp[3] = wp[3] ^ (1 << 31)
        // wp[4] = wp[4] ^ (1 << 31)
        // flip the 1 bit (second least significant bit) in 20th word 
        // BEST HERE!!
        // wp[20] = wp[20] ^ (1 << 1)
        // wp[21] = wp[21] ^ (1 << 6)
        // wp[22] = wp[22] ^ (1 << 1)
        // wp[23] = wp[23] ^ (1 << 31)
        // wp[24] = wp[24] ^ (1 << 31)
        // flip the 1 bit (second least significant bit) in 40th word 
        // wp[40] = wp[40] ^ (1 << 1)
        // wp[41] = wp[41] ^ (1 << 6)
        // wp[42] = wp[42] ^ (1 << 1)
        // wp[43] = wp[43] ^ (1 << 31)
        // wp[44] = wp[44] ^ (1 << 31)
        // flip the 1 bit (second least significant bit) in 60th word 
        // wp[60] = wp[60] ^ (1 << 1)
        // wp[61] = wp[61] ^ (1 << 6)
        // wp[62] = wp[62] ^ (1 << 1)
        // wp[63] = wp[63] ^ (1 << 31)
        // wp[64] = wp[64] ^ (1 << 31)

        for i := 0; i < 5; i++ {
            tState[i] = state[i]
            tpState[i] = pState[i]
        }
        
        for i := 0; i < 58; i++ {
            tState = round(i, tState, w[i])
            tpState = round(i, tpState, wp[i])
            if (!test(tState, tpState, i)){
                fmt.Printf("round %d:\n", i)
                fmt.Printf("\tA  = 0x%08x,\t%032b\n", tState[0], tState[0])
                fmt.Printf("\tA' = 0x%08x,\t%032b\tBit differences = %d\n", 
                    tpState[0], tpState[0], 
                    countBitDifferences(tState[0], tpState[0]))
                fmt.Printf("\tB  = 0x%08x,\t%032b\n", tState[1], tState[1])
                fmt.Printf("\tB' = 0x%08x,\t%032b\tBit differences = %d\n", 
                    tpState[1], tpState[1], 
                    countBitDifferences(tState[1], tpState[1]))
                fmt.Printf("\tC  = 0x%08x,\t%032b\n", tState[2], tState[2])
                fmt.Printf("\tC' = 0x%08x,\t%032b\tCit differences = %d\n", 
                    tpState[2], tpState[2], 
                    countBitDifferences(tState[2], tpState[2]))
                fmt.Printf("\tD  = 0x%08x,\t%032b\n", tState[3], tState[3])
                fmt.Printf("\tD' = 0x%08x,\t%032b\tDit differences = %d\n", 
                    tpState[3], tpState[3], 
                    countBitDifferences(tState[3], tpState[3]))
                fmt.Printf("\tE  = 0x%08x,\t%032b\n", tState[4], tState[4])
                fmt.Printf("\tE' = 0x%08x,\t%032b\tEit differences = %d\n", 
                    tpState[4], tpState[4], 
                    countBitDifferences(tState[4], tpState[4]))
            }
        }


        state[0] =  (state[0] + tState[0]) & 0xffffffff
        state[1] =  (state[1] + tState[1]) & 0xffffffff
        state[2] =  (state[2] + tState[2]) & 0xffffffff
        state[3] =  (state[3] + tState[3]) & 0xffffffff
        state[4] =  (state[4] + tState[4]) & 0xffffffff
        pState[0] =  (pState[0] + tpState[0]) & 0xffffffff
        pState[1] =  (pState[1] + tpState[1]) & 0xffffffff
        pState[2] =  (pState[2] + tpState[2]) & 0xffffffff
        pState[3] =  (pState[3] + tpState[3]) & 0xffffffff
        pState[4] =  (pState[4] + tpState[4]) & 0xffffffff
        // fmt.Println("new chunk:")
        // fmt.Printf("\tA  = 0x%08x\n", state[0])
        // fmt.Printf("\tA' = 0x%08x\n", pState[0])
        // fmt.Printf("\tB  = 0x%08x\n", state[1])
        // fmt.Printf("\tB' = 0x%08x\n", pState[1])
        // fmt.Printf("\tC  = 0x%08x\n", state[2])
        // fmt.Printf("\tC' = 0x%08x\n", pState[2])
        // fmt.Printf("\tD  = 0x%08x\n", state[3])
        // fmt.Printf("\tD' = 0x%08x\n", pState[3])
        // fmt.Printf("\tE  = 0x%08x\n", state[4])
        // fmt.Printf("\tE' = 0x%08x\n", pState[4])
	}

    fmt.Printf("sha-1 digest: %08x%08x%08x%08x%08x\n", state[0], state[1],
        state[2], state[3], state[4])
    fmt.Printf("sha-1 digest: %08x%08x%08x%08x%08x\n", pState[0], pState[1],
        pState[2], pState[3], pState[4])
}