// Copyright 2014 Brian J. Downs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Simple library for working with passwords in Go.
// All generated passwords are going to be a minimum of 8
// characters in length.
package GoPasswordUtilities

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	characters    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+,.?/:;{}[]~"
	wordsLocation = "/usr/share/dict/words"
)

var (
	passwordScores = map[int]string{
		0: "Horrible",
		1: "Weak",
		2: "Medium",
		3: "Strong",
		4: "Very Strong"}
)

type Password struct {
	Pass            string
	Length          int
	Score           int
	ContainsUpper   bool
	ContainsLower   bool
	ContainsNumber  bool
	ContainsSpecial bool
	DictionaryBased bool
}

type SaltConf struct {
	Length int
}

// New is used when a user enters a password as well as the
// being called from the GeneratePassword function.
func New(password string) *Password {
	return &Password{Pass: password, Length: len(password)}
}

// GeneratePassword will generate and return a password as a string and as a
// byte slice of the given length.
func GeneratePassword(length int) *Password {
	passwordBuffer := new(bytes.Buffer)
	randBytes := make([]byte, length)

	if _, err := rand.Read(randBytes); err == nil {
		for j := 0; j < length; j++ {
			tmpIndex := int(randBytes[j]) % len(characters)
			char := characters[tmpIndex]
			passwordBuffer.WriteString(string(char))
		}
	}
	return New(passwordBuffer.String())
}

// GenerateVeryStrongPassword will generate a "Very Strong" password.
func GenerateVeryStrongPassword(length int) *Password {
	for {
		p := GeneratePassword(length)
		p.ProcessPassword()
		if p.Score == 4 {
			return p
		}
	}
}

// getRandomBytes will generate random bytes.  This is for internal
// use in the library itself.
func getRandomBytes(length int) []byte {
	randomData := make([]byte, length)
	if _, err := rand.Read(randomData); err != nil {
		log.Fatalf("%v\n", err)
	}
	return randomData
}

// MD5 sum for the given password.  If a SaltConf
// pointer is given as a parameter a salt with the given
// length will be returned with it included in the hash.
func (p *Password) MD5(saltConf ...*SaltConf) ([16]byte, []byte) {
	if len(saltConf) > 0 {
		var saltLength int

		for _, i := range saltConf[0:] {
			saltLength = i.Length
		}

		salt := getRandomBytes(saltLength)
		return md5.Sum([]byte(fmt.Sprintf("%s%x", p.Pass, salt))), salt
	}
	return md5.Sum([]byte(p.Pass)), nil
}

// SHA256 sum for the given password.  If a SaltConf
// pointer is given as a parameter a salt with the given
// length will be returned with it included in the hash.
func (p *Password) SHA256(saltConf ...*SaltConf) ([32]byte, []byte) {
	if len(saltConf) > 0 {
		var saltLength int

		for _, i := range saltConf[0:] {
			saltLength = i.Length
		}

		salt := getRandomBytes(saltLength)
		return sha256.Sum256([]byte(fmt.Sprintf("%s%x", p.Pass, salt))), salt
	}
	return sha256.Sum256([]byte(p.Pass)), nil
}

// SHA512 sum for the given password.  If a SaltConf
// pointer is given as a parameter a salt with the given
// length will be returned with it included in the hash.
func (p *Password) SHA512(saltConf ...*SaltConf) ([64]byte, []byte) {
	if len(saltConf) > 0 {
		var saltLength int

		for _, i := range saltConf[0:] {
			saltLength = i.Length
		}

		salt := getRandomBytes(saltLength)
		return sha512.Sum512([]byte(fmt.Sprintf("%s%x", p.Pass, salt))), salt
	}
	return sha512.Sum512([]byte(p.Pass)), nil
}

// GetLength will provide the length of the password.  This method is
// being put on the password struct in case someone decides not to
// do a complexity check.
func (p *Password) GetLength() int {
	return p.Length
}

// ProcessPassword will parse the password and populate the Password struct attributes.
func (p *Password) ProcessPassword() {
	matchLower := regexp.MustCompile(`[a-z]`)
	matchUpper := regexp.MustCompile(`[A-Z]`)
	matchNumber := regexp.MustCompile(`[0-9]`)
	matchSpecial := regexp.MustCompile(`[\!\@\#\$\%\^\&\*\(\\\)\-_\=\+\,\.\?\/\:\;\{\}\[\]~]`)

	if p.Length < 8 {
		log.Fatalln("password isn't long enough for evaluation")
	}

	if matchLower.MatchString(p.Pass) {
		p.ContainsLower = true
		p.Score++
	}
	if matchUpper.MatchString(p.Pass) {
		p.ContainsUpper = true
		p.Score++
	}
	if matchNumber.MatchString(p.Pass) {
		p.ContainsNumber = true
		p.Score++
	}
	if matchSpecial.MatchString(p.Pass) {
		p.ContainsSpecial = true
		p.Score++
	}
	if searchDict(p.Pass) {
		p.DictionaryBased = true
		p.Score--
	}
}

// searchDict will search the words list for an occurance of the
// given word.  Requires wamerican || wbritish || wordlist || words
// to be installed.
func searchDict(word string) bool {
	file, err := os.Open(wordsLocation)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(strings.ToLower(scanner.Text()), word) {
			return true
		}
	}
	return false
}

// GetScore will provide the score of the password.
func (p *Password) GetScore() int {
	return p.Score
}

// HasUpper indicates whether the password contains an upper case letter.
func (p *Password) HasUpper() bool {
	return p.ContainsUpper
}

// HasLower indicates whether the password contains a lower case letter.
func (p *Password) HasLower() bool {
	return p.ContainsLower
}

// HasNumber indicates whether the password contains a number.
func (p *Password) HasNumber() bool {
	return p.ContainsNumber
}

// HasSpecial indicates whether the password contains a special character.
func (p *Password) HasSpecial() bool {
	return p.ContainsSpecial
}

// ComplexityRating provides the rating for the password.
func (p *Password) ComplexityRating() string {
	return passwordScores[p.Score]
}

// InDictionary will return true or false if it's been detected
// that the given password is a dictionary based.
func (p *Password) InDictionary() bool {
	return p.DictionaryBased
}
