// ADVOCATE-FILE-START

package runtime

/*
 * Get a string representation of an uint64
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func uint64ToString(n uint64) string {
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return uint64ToString(n/10) + string(rune(n%10+'0'))
	}
}

/*
 * Get a string representation of an int64
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func int64ToString(n int64) string {
	if n < 0 {
		return "-" + int64ToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return int64ToString(n/10) + string(rune(n%10+'0'))

	}
}

/*
 * Get a string representation of an int32
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func int32ToString(n int32) string {
	if n < 0 {
		return "-" + int32ToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return int32ToString(n/10) + string(rune(n%10+'0'))

	}
}

/*
 * Get a string representation of an uint32
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func uint32ToString(n uint32) string {
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return uint32ToString(n/10) + string(rune(n%10+'0'))
	}
}

/*
 * Get a string representation of an int
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func intToString(n int) string {
	if n < 0 {
		return "-" + intToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return intToString(n/10) + string(rune(n%10+'0'))

	}
}

var advocateCurrentRoutineID uint64
var advocateCurrentObjectID uint64
var advocateGlobalCounter uint64

var advocateCurrentRoutineIDMutex mutex
var advocateCurrentObjectIDMutex mutex
var advocateGlobalCounterMutex mutex

/*
 * Get a new id for a routine
 * Return:
 * 	new id
 */
func GetAdvocateRoutineID() uint64 {
	lock(&advocateCurrentRoutineIDMutex)
	defer unlock(&advocateCurrentRoutineIDMutex)
	advocateCurrentRoutineID++
	return advocateCurrentRoutineID
}

/*
 * Get a new id for a mutex, channel or waitgroup
 * Return:
 * 	new id
 */
func GetAdvocateObjectID() uint64 {
	lock(&advocateCurrentObjectIDMutex)
	defer unlock(&advocateCurrentObjectIDMutex)
	advocateCurrentObjectID++
	return advocateCurrentObjectID
}

/*
 * Get a new counter
 * Return:
 * 	new counter value
 */
func GetAdvocateCounter() uint64 {
	lock(&advocateGlobalCounterMutex)
	defer unlock(&advocateGlobalCounterMutex)
	advocateGlobalCounter++
	return advocateGlobalCounter
}

/*
 * Split a string by the seperator
 */
func splitString(line string, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(line); i++ {
		if line[i] == sep[0] {
			result = append(result, line[start:i])
			start = i + 1
		}
	}
	result = append(result, line[start:])
	return result
}

/*
 * Convert a string to an integer
 * Works only with positive integers
 */
func stringToInt(s string) int {
	var result int
	sign := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '-' && i == 0 {
			sign = -1
		} else if s[i] >= '0' && s[i] <= '9' {
			result = result*10 + int(s[i]-'0')
		} else {
			panic("Invalid input")
		}
	}
	return result * sign
}

/*
 * Check if a list of integers contains an element
 * Args:
 * 	list: list of integers
 * 	elem: element to check
 * Return:
 * 	true if the list contains the element, false otherwise
 */
func containsInt(list []int, elem int) bool {
	for _, e := range list {
		if e == elem {
			return true
		}
	}
	return false
}

/*
 * Slow down the execution of the program
 */
func slowExecution() {
	for i := 0; i < 1e8; i++ {
		// do nothing
	}
}

// ADVOCATE-FILE-END
