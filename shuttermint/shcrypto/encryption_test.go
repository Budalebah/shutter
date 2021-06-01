package shcrypto

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"gotest.tools/v3/assert"
)

func TestRandomSigma(t *testing.T) {
	firstByteAlways0 := true
	lastByteAlways0 := true
	for i := 0; i < 10; i++ {
		sigma, err := RandomSigma(rand.Reader)
		assert.NilError(t, err)
		if sigma[0] != 0 {
			firstByteAlways0 = false
		}
		if sigma[BlockSize-1] != 0 {
			lastByteAlways0 = false
		}
	}
	assert.Assert(t, !firstByteAlways0)
	assert.Assert(t, !lastByteAlways0)
}

func TestPadding(t *testing.T) {
	testCases := []struct {
		mHex  string
		bsHex []string
	}{
		{
			"",
			[]string{
				strings.Repeat("20", 32),
			},
		},
		{
			"99",
			[]string{
				"99" + strings.Repeat("1f", 31),
			},
		},
		{
			"9999",
			[]string{
				"9999" + strings.Repeat("1e", 30),
			},
		},
		{
			strings.Repeat("99", 31),
			[]string{
				strings.Repeat("99", 31) + "01",
			},
		},
		{
			strings.Repeat("99", 32),
			[]string{
				strings.Repeat("99", 32),
				strings.Repeat("20", 32),
			},
		},
		{
			strings.Repeat("99", 33),
			[]string{
				strings.Repeat("99", 32),
				"99" + strings.Repeat("1f", 31),
			},
		},
	}

	for _, test := range testCases {
		m, err := hex.DecodeString(test.mHex)
		assert.NilError(t, err)
		bs := []Block{}
		for _, bHex := range test.bsHex {
			bBytes, err := hex.DecodeString(bHex)
			assert.NilError(t, err)
			assert.Equal(t, 32, len(bBytes))
			var bByteArray [32]byte
			copy(bByteArray[:], bBytes)
			b := Block(bByteArray)
			bs = append(bs, b)
		}

		padded := PadMessage(m)
		assert.DeepEqual(t, bs, padded)
	}
}

func TestUnpadding(t *testing.T) {
	invalidBs := [][]string{
		{
			strings.Repeat("00", 32),
		},
		{
			"aabbcc" + strings.Repeat("00", 29),
		},
		{
			strings.Repeat("21", 32),
		},
		{
			strings.Repeat("99", 32),
			strings.Repeat("00", 32),
		},
	}
	for _, bsHex := range invalidBs {
		bs := []Block{}
		for _, bHex := range bsHex {
			bBytes, err := hex.DecodeString(bHex)
			assert.NilError(t, err)
			assert.Equal(t, 32, len(bBytes))
			var bByteArray [32]byte
			copy(bByteArray[:], bBytes)
			b := Block(bByteArray)
			bs = append(bs, b)
		}

		_, err := UnpadMessage(bs)
		assert.Assert(t, err != nil)
	}

	testCases := []struct {
		bsHex []string
		mHex  string
	}{
		{
			[]string{
				strings.Repeat("20", 32),
			},
			"",
		},
		{
			[]string{
				"99" + strings.Repeat("1f", 31),
			},
			"99",
		},
		{
			[]string{
				strings.Repeat("99", 31) + "01",
			},
			strings.Repeat("99", 31),
		},
		{
			[]string{
				strings.Repeat("99", 32),
				strings.Repeat("20", 32),
			},
			strings.Repeat("99", 32),
		},
	}
	for _, test := range testCases {
		bs := []Block{}
		for _, bHex := range test.bsHex {
			bBytes, err := hex.DecodeString(bHex)
			assert.NilError(t, err)
			assert.Equal(t, 32, len(bBytes))
			var bByteArray [32]byte
			copy(bByteArray[:], bBytes)
			b := Block(bByteArray)
			bs = append(bs, b)
		}
		m, err := hex.DecodeString(test.mHex)
		assert.NilError(t, err)

		unpadded, err := UnpadMessage(bs)
		assert.NilError(t, err)
		assert.DeepEqual(t, m, unpadded)
	}
}

func TestPaddingRoundtrip(t *testing.T) {
	ms := [][]byte{
		[]byte(""),
		[]byte("a"),
		bytes.Repeat([]byte("a"), 31),
		bytes.Repeat([]byte("a"), 32),
		bytes.Repeat([]byte("a"), 33),
	}
	for i := 0; i < 100; i++ {
		l, err := rand.Int(rand.Reader, big.NewInt(100))
		assert.NilError(t, err)
		m := make([]byte, l.Int64())
		_, err = rand.Read(m)
		assert.NilError(t, err)
		ms = append(ms, m)
	}
	for _, m := range ms {
		padded := PadMessage(m)
		unpadded, err := UnpadMessage(padded)
		assert.NilError(t, err)
		assert.DeepEqual(t, m, unpadded)
	}
}

func TestRoundTrip(t *testing.T) {
	// first generate keys
	n := 3
	threshold := uint64(2)
	_ = ComputeEpochID(uint64(10))
	epochID := (*EpochID)(new(bn256.G1).ScalarBaseMult(big.NewInt(1)))

	ps := []*Polynomial{}
	gammas := []*Gammas{}
	for i := 0; i < n; i++ {
		p, err := RandomPolynomial(rand.Reader, threshold-1)
		assert.NilError(t, err)
		ps = append(ps, p)
		gammas = append(gammas, p.Gammas())
	}

	eonSecretKeyShares := []*EonSecretKeyShare{}
	epochSecretKeyShares := []*EpochSecretKeyShare{}
	eonSecretKey := big.NewInt(0)
	for i := 0; i < n; i++ {
		eonSecretKey.Add(eonSecretKey, ps[i].Eval(big.NewInt(0)))

		ss := []*big.Int{}
		for j := 0; j < n; j++ {
			s := ps[j].EvalForKeyper(i)
			ss = append(ss, s)
		}
		eonSecretKeyShares = append(eonSecretKeyShares, ComputeEonSecretKeyShare(ss))
		_ = ComputeEonPublicKeyShare(i, gammas)
		epochSecretKeyShares = append(epochSecretKeyShares, ComputeEpochSecretKeyShare(eonSecretKeyShares[i], epochID))
	}
	eonPublicKey := ComputeEonPublicKey(gammas)
	assert.DeepEqual(t, new(bn256.G2).ScalarBaseMult(eonSecretKey), (*bn256.G2)(eonPublicKey), G2Comparer)
	epochSecretKey, err := ComputeEpochSecretKey(
		[]int{0, 1},
		[]*EpochSecretKeyShare{epochSecretKeyShares[0], epochSecretKeyShares[1]},
		threshold)
	assert.NilError(t, err)

	// now encrypt and decrypt message
	m := []byte("hello")
	sigma, err := RandomSigma(rand.Reader)
	assert.NilError(t, err)

	encM := Encrypt(m, eonPublicKey, epochID, sigma)
	decM, err := encM.Decrypt(epochSecretKey)
	assert.NilError(t, err)
	assert.DeepEqual(t, m, decM)
}
