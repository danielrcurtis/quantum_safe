package main

import (
	"errors"
	"fmt"
	"quantum_safe/logger"
)

var PublicKey = Matrix{
	{47, -77, -85},
	{-49, 78, 50},
	{57, -78, 99},
}

type Matrix [3][3]float64

type Vector struct {
	X, Y, Z float64
}

var r = Vector{} //{-4365, 4923, -5772}

func main() {
	logger.Log.Sugar().Infof("Public Key: %+v", PublicKey)
	// Inverse of the PublicKey
	inversePublicKey, err := Inverse3x3(PublicKey)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	logger.Log.Sugar().Infof("Inverse Public Key: %+v", inversePublicKey)

	// Define ASCII values of 'H', 'T', and 'B'
	expectedChars := []rune{'H', 'T', 'B'}

	r, err := FindCorrectRParallel(PublicKey, inversePublicKey, expectedChars)
	if err != nil {
		logger.Log.Sugar().Errorf("Error finding r: %v", err)
		return
	}

	logger.Log.Sugar().Infof("Found 'r' value: %+v", r)

	// Decrypt each vector in Flag and build the message
	var message string
	for _, encryptedVector := range Flag {
		logger.Log.Sugar().Infof("Encrypted Vector: %+v", encryptedVector)
		decryptedChar := DecryptCharacter(encryptedVector, inversePublicKey)
		logger.Log.Sugar().Infof("Decrypted Character: %v", decryptedChar)
		if decryptedChar != 0 {
			message += string(decryptedChar)
		}
	}

	// Print the decrypted message
	fmt.Println("Decrypted Message: ", message)
}

func FindCorrectRParallel(pubKey Matrix, invPubKey Matrix, expectedChars []rune) (Vector, error) {
	results := make(chan Vector)
	errChan := make(chan error)
	done := make(chan bool)
	workerCount := 8 // Number of workers; adjust based on your CPU cores

	// Calculate the range each worker should handle
	rangePerWorker := 40000 / workerCount // Assuming the range is from -10000 to +10000

	for i := 0; i < workerCount; i++ {
		start := -10000 + i*rangePerWorker
		end := start + rangePerWorker

		// Start a worker goroutine
		go worker(pubKey, invPubKey, expectedChars, start, end, results, errChan, done)
	}

	for i := 0; i < workerCount; i++ {
		select {
		case result := <-results:
			close(done) // Signal other workers to stop
			return result, nil
		case err := <-errChan:
			return Vector{}, err
		}
	}

	return Vector{}, errors.New("no matching r found")
}

func worker(pubKey Matrix, invPubKey Matrix, expectedChars []rune, start, end int, results chan<- Vector, errChan chan<- error, done <-chan bool) {
	for x := start; x < end; x++ {
		select {
		case <-done:
			return // Stop if done signal is received
		default:
			for y := -10000; y <= 10000; y++ {
				for z := -10000; z <= 10000; z++ {
					testR := Vector{float64(x), float64(y), float64(z)}
					if CheckR(pubKey, invPubKey, testR, expectedChars) {
						results <- testR
						return
					}
				}
			}
		}
	}
	errChan <- errors.New("range completed without success")
}

func FindCorrectR(pubKey Matrix, invPubKey Matrix, expectedChars []rune) (Vector, error) {
	for x := -10000; x <= 10000; x++ {
		for y := -10000; y <= 10000; y++ {
			for z := -10000; z <= 10000; z++ {
				testR := Vector{float64(x), float64(y), float64(z)}
				if CheckR(pubKey, invPubKey, testR, expectedChars) {
					return testR, nil
				}
			}
		}
	}
	return Vector{}, errors.New("no matching r found")
}

func CheckR(pubKey Matrix, invPubKey Matrix, testR Vector, expectedChars []rune) bool {
	for i, char := range expectedChars {
		decryptedChar := DecryptCharacterWithR(Flag[i], invPubKey, testR)
		if decryptedChar != char {
			return false
		}
	}
	return true
}

func DecryptCharacterWithR(encryptedVector Vector, inversePublicKey Matrix, r Vector) rune {
	// Adjust the vector with assumed random integers
	subtractedVector := SubtractVectors(encryptedVector, r)

	// Multiply by the inverse of the PublicKey
	decryptedVector := MultiplyMatrixVector(inversePublicKey, subtractedVector)

	// Check if the X component is a printable ASCII character
	if isPrintableASCII(int(decryptedVector.X)) {
		return rune(int(decryptedVector.X))
	}
	return 0
}

// DecryptCharacter attempts to brute-force decrypt the character
func DecryptCharacter(encryptedVector Vector, inversePublicKey Matrix) rune {
	for i := 0; i <= 100; i++ {
		for j := 0; j <= 100; j++ {
			// Adjust the vector with assumed random integers
			adjustedVector := Vector{X: encryptedVector.X, Y: float64(i), Z: float64(j)}

			// Subtract r from the adjusted vector
			subtractedVector := SubtractVectors(adjustedVector, r)

			// Multiply by the inverse of the PublicKey
			decryptedVector := MultiplyMatrixVector(inversePublicKey, subtractedVector)

			// Check if the X component is a printable ASCII character
			if isPrintableASCII(int(decryptedVector.X)) {
				logger.Log.Sugar().Infof("Decrypted Vector: %+v", decryptedVector)
				return rune(int(decryptedVector.X))
			} else {
				if isPrintableASCII(int(decryptedVector.Y)) {
					logger.Log.Sugar().Infof("Decrypted Vector: %+v", decryptedVector)
					return rune(int(decryptedVector.Y))
				}
			}
		}
	}
	return 0
}

// Check if a value is a printable ASCII character
func isPrintableASCII(value int) bool {
	return value >= 32 && value <= 126
}

// MultiplyMatrixVector multiplies a matrix by a vector.
func MultiplyMatrixVector(m Matrix, v Vector) Vector {
	var result Vector
	result.X = m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z
	result.Y = m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z
	result.Z = m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z
	return result
}

// Inverse3x3 computes the inverse of a 3x3 matrix.
func Inverse3x3(m Matrix) (Matrix, error) {
	// Calculate the determinant of the matrix
	det := m[0][0]*(m[1][1]*m[2][2]-m[1][2]*m[2][1]) -
		m[0][1]*(m[1][0]*m[2][2]-m[1][2]*m[2][0]) +
		m[0][2]*(m[1][0]*m[2][1]-m[1][1]*m[2][0])

	logger.Log.Sugar().Infof("Determinant of matrix: %v", det)

	// Check if the determinant is zero (matrix is not invertible)
	if det == 0 {
		return Matrix{}, errors.New("matrix is not invertible, determinant is zero")
	}

	// Compute the inverse matrix using floating-point division
	invDet := 1.0 / det
	var inverse Matrix
	inverse[0][0] = (m[1][1]*m[2][2] - m[1][2]*m[2][1]) * invDet
	inverse[0][1] = (m[0][2]*m[2][1] - m[0][1]*m[2][2]) * invDet
	inverse[0][2] = (m[0][1]*m[1][2] - m[0][2]*m[1][1]) * invDet
	inverse[1][0] = (m[1][2]*m[2][0] - m[1][0]*m[2][2]) * invDet
	inverse[1][1] = (m[0][0]*m[2][2] - m[0][2]*m[2][0]) * invDet
	inverse[1][2] = (m[0][2]*m[1][0] - m[0][0]*m[1][2]) * invDet
	inverse[2][0] = (m[1][0]*m[2][1] - m[1][1]*m[2][0]) * invDet
	inverse[2][1] = (m[0][1]*m[2][0] - m[0][0]*m[2][1]) * invDet
	inverse[2][2] = (m[0][0]*m[1][1] - m[0][1]*m[1][0]) * invDet

	return inverse, nil
}

// SubtractVectors subtracts one vector from another.
func SubtractVectors(a, b Vector) Vector {
	return Vector{
		X: a.X - b.X,
		Y: a.Y - b.Y,
		Z: a.Z - b.Z,
	}
}
