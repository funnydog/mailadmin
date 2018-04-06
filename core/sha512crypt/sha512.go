package sha512crypt

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"strconv"
)

const (
	magicPrefix   = '6'
	maxSaltSize   = 16
	minRounds     = 1000
	maxRounds     = 999999999
	defaultRounds = 5000
	alphabet      = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var ErrMismatchedHashAndPassword = errors.New("sha512crypt: hashedPassword is not the hash of the given password")
var ErrHashTooShort = errors.New("sha512crypt: hashedSecret too short to be a sha512crypted password")

type InvalidHashVersionError byte

func (ih InvalidHashVersionError) Error() string {
	return fmt.Sprintf("sha512crypt: crypt algorithm version must be '%c', but hashedSecret has '%c' instead'", magicPrefix, ih)
}

type InvalidHashPrefixError byte

func (ih InvalidHashPrefixError) Error() string {
	return fmt.Sprintf("sha512crypt: hashes must start with '$', but hashedSecret started with '%c'", byte(ih))
}

type InvalidSaltPrefixError byte

func (is InvalidSaltPrefixError) Error() string {
	return fmt.Sprintf("sha512crypt: salt must start with '$', but hashedSecred salt started with '%c'", byte(is))
}

type InvalidRoundsError int

func (ir InvalidRoundsError) Error() string {
	return fmt.Sprintf("sha512crypt: rounds %d is outside allowed range (%d,%d)", int(ir), int(minRounds), int(maxRounds))
}

var roundsPrefix = []byte("rounds=")

type hashed struct {
	hash   []byte
	salt   []byte
	rounds int
}

func (p *hashed) Hash() []byte {
	arr := make([]byte, 123)
	arr[0] = '$'
	arr[1] = magicPrefix
	arr[2] = '$'
	n := 3

	if p.rounds != defaultRounds {
		rounds := []byte(fmt.Sprintf("rounds=%d$", p.rounds))
		copy(arr[n:], rounds)
		n += len(rounds)
	}

	copy(arr[n:], p.salt)
	n += len(p.salt)
	arr[n] = '$'
	n++
	copy(arr[n:], p.hash)
	n += len(p.hash)
	return arr[:n]
}

func (p *hashed) decodeVersion(hash []byte) (int, error) {
	if hash[0] != '$' {
		return 0, InvalidHashPrefixError(hash[0])
	}

	if hash[1] != magicPrefix {
		return 0, InvalidHashVersionError(hash[1])
	}

	if hash[2] != '$' {
		return 0, InvalidSaltPrefixError(hash[2])
	}

	return 3, nil
}

func (p *hashed) decodeRounds(hash []byte) (int, error) {
	if !bytes.HasPrefix(hash, roundsPrefix) {
		p.rounds = defaultRounds
		return 0, nil
	}

	end := bytes.IndexByte(hash, '$')
	if end == -1 {
		return 0, ErrHashTooShort
	}

	rounds, err := strconv.ParseInt(string(hash[7:end]), 10, 32)
	if err != nil {
		return 0, ErrHashTooShort
	}

	if rounds < minRounds {
		return 0, InvalidRoundsError(int(rounds))
	}

	if rounds > maxRounds {
		return 0, InvalidRoundsError(int(rounds))
	}

	p.rounds = int(rounds)
	return end + 1, nil
}

func GenerateFromPassword(password []byte) ([]byte, error) {
	p, err := newFromPassword(password)
	if err != nil {
		return nil, err
	}

	return p.Hash(), nil
}

func CompareHashAndPassword(hashedPassword, password []byte) error {
	p, err := newFromHash(hashedPassword)
	if err != nil {
		return err
	}

	otherHash, err := sha512crypt(password, p.salt, p.rounds)
	if err != nil {
		return err
	}

	otherP := &hashed{otherHash, p.salt, p.rounds}
	if subtle.ConstantTimeCompare(p.Hash(), otherP.Hash()) == 1 {
		return nil
	}

	return ErrMismatchedHashAndPassword
}

func base64Encode(src []byte) []byte {
	hashSize := (len(src)*8 + 5) / 6
	hash := make([]byte, hashSize)

	srclen := len(src)
	j := 0
	for i := 0; i < srclen; {
		var w uint
		var count int
		switch srclen - i {
		default:
			w = (uint(src[i+2]) << 16) | (uint(src[i+1]))<<8 | uint(src[i])
			i += 3
			count = 4
		case 2:
			w = (uint(src[i+1]))<<8 | uint(src[i])
			i += 2
			count = 3
		case 1:
			w = uint(src[i])
			i += 1
			count = 2
		}

		for ; count > 0; count-- {
			hash[j] = alphabet[w&0x3F]
			w >>= 6
			j++
		}
	}
	return hash
}

func newFromPassword(password []byte) (*hashed, error) {
	p := new(hashed)

	unencodedSalt := make([]byte, maxSaltSize)
	_, err := io.ReadFull(rand.Reader, unencodedSalt)
	if err != nil {
		return nil, err
	}

	p.salt = base64Encode(unencodedSalt)[:maxSaltSize]
	p.rounds = defaultRounds
	hash, err := sha512crypt(password, p.salt, p.rounds)
	if err != nil {
		return nil, err
	}
	p.hash = hash
	return p, err
}

func newFromHash(hashedSecret []byte) (*hashed, error) {
	p := new(hashed)
	n, err := p.decodeVersion(hashedSecret)
	if err != nil {
		return nil, err
	}

	hashedSecret = hashedSecret[n:]
	n, err = p.decodeRounds(hashedSecret)
	if err != nil {
		return nil, err
	}
	hashedSecret = hashedSecret[n:]

	end := bytes.IndexByte(hashedSecret, '$')
	if end == -1 {
		return nil, ErrHashTooShort
	}

	p.salt = make([]byte, 16, 16+2)
	copy(p.salt, hashedSecret[:end])

	hashedSecret = hashedSecret[end+1:]
	p.hash = make([]byte, len(hashedSecret))
	copy(p.hash, hashedSecret)

	return p, nil
}
func sha512crypt(password, salt []byte, rounds int) ([]byte, error) {
	// digest B
	B := sha512.New()
	B.Write(password)
	B.Write(salt)
	B.Write(password)
	BSum := B.Sum(nil)

	// digest A
	A := sha512.New()
	A.Write(password)
	A.Write(salt)

	var i int
	for i = len(password); i > 64; i -= 64 {
		A.Write(BSum)
	}
	A.Write(BSum[:i])

	for i = len(password); i > 0; i >>= 1 {
		if (i & 1) != 0 {
			A.Write(BSum)
		} else {
			A.Write(password)
		}
	}
	ASum := A.Sum(nil)

	// digest DP
	DP := sha512.New()
	for i = 0; i < len(password); i++ {
		DP.Write(password)
	}
	DPSum := DP.Sum(nil)

	DPSeq := make([]byte, 0, len(password))
	for i = len(password); i > 64; i -= 64 {
		DPSeq = append(DPSeq, DPSum...)
	}
	DPSeq = append(DPSeq, DPSum[:i]...)

	// digest DS
	DS := sha512.New()
	for i := 0; i < (16 + int(ASum[0])); i++ {
		DS.Write(salt)
	}
	DSSum := DS.Sum(nil)

	DSSeq := make([]byte, 0, len(salt))
	for i = len(salt); i > 64; i -= 64 {
		DSSeq = append(DSSeq, DSSum...)
	}
	DSSeq = append(DSSeq, DSSum[:i]...)

	CSum := ASum
	for i = 0; i < rounds; i++ {
		C := sha512.New()

		if (i & 1) != 0 {
			C.Write(DPSeq)
		} else {
			C.Write(CSum)
		}

		if (i % 3) != 0 {
			C.Write(DSSeq)
		}

		if (i % 7) != 0 {
			C.Write(DPSeq)
		}

		if (i & 1) != 0 {
			C.Write(CSum)
		} else {
			C.Write(DPSeq)
		}

		CSum = C.Sum(nil)
	}

	out := base64Encode([]byte{
		CSum[42], CSum[21], CSum[0],
		CSum[1], CSum[43], CSum[22],
		CSum[23], CSum[2], CSum[44],
		CSum[45], CSum[24], CSum[3],
		CSum[4], CSum[46], CSum[25],
		CSum[26], CSum[5], CSum[47],
		CSum[48], CSum[27], CSum[6],
		CSum[7], CSum[49], CSum[28],
		CSum[29], CSum[8], CSum[50],
		CSum[51], CSum[30], CSum[9],
		CSum[10], CSum[52], CSum[31],
		CSum[32], CSum[11], CSum[53],
		CSum[54], CSum[33], CSum[12],
		CSum[13], CSum[55], CSum[34],
		CSum[35], CSum[14], CSum[56],
		CSum[57], CSum[36], CSum[15],
		CSum[16], CSum[58], CSum[37],
		CSum[38], CSum[17], CSum[59],
		CSum[60], CSum[39], CSum[18],
		CSum[19], CSum[61], CSum[40],
		CSum[41], CSum[20], CSum[62],
		CSum[63],
	})

	A.Reset()
	B.Reset()
	DP.Reset()
	for i = 0; i < len(ASum); i++ {
		ASum[i] = 0
	}
	for i = 0; i < len(BSum); i++ {
		BSum[i] = 0
	}
	for i = 0; i < len(DPSeq); i++ {
		DPSeq[i] = 0
	}

	return out, nil
}
