package analysis

import (
	"analyzer/logging"
	"strconv"
)

/*
 * Add a lock to the lockSet of a routine. Also save the vector clock of the acquire
 * Args:
 *   routine (int): The routine id
 *   lock (int): The id of the mutex
 *   tId (string): The trace id of the mutex operation
 *   vc (VectorClock): The current vector clock
 */
func lockSetAddLock(routine int, lock int, tID string, vc VectorClock) {
	if _, ok := lockSet[routine]; !ok {
		lockSet[routine] = make(map[int]string)
	}
	if _, ok := mostRecentAcquire[routine]; !ok {
		mostRecentAcquire[routine] = make(map[int]VectorClockTID)
	}

	if posOld, ok := lockSet[routine][lock]; ok {
		errorMsg := "Lock " + strconv.Itoa(lock) +
			" already in lockSet for routine " + strconv.Itoa(routine)
		logging.Debug(errorMsg, logging.ERROR)

		// this is a double locking
		found := "Double locking:\n"
		found += "\tlock1: " + posOld + "\n"
		found += "\tlock2: " + tID
		logging.Result(found, logging.CRITICAL)
	}

	lockSet[routine][lock] = tID
	mostRecentAcquire[routine][lock] = VectorClockTID{vc, tID}
}

/*
 * Remove a lock from the lockSet of a routine
 * Args:
 *   routine (int): The routine id
 *   lock (int): The id of the mutex
 */
func lockSetRemoveLock(routine int, lock int) {
	if _, ok := lockSet[routine][lock]; !ok {
		errorMsg := "Lock " + strconv.Itoa(lock) +
			" not in lockSet for routine " + strconv.Itoa(routine)
		logging.Debug(errorMsg, logging.ERROR)
		return
	}
	delete(lockSet[routine], lock)
}

func checkForMixedDeadlock(routineSend int, routineRevc int) {
	for m := range lockSet[routineSend] {
		_, ok1 := mostRecentAcquire[routineRevc][m]
		_, ok2 := mostRecentAcquire[routineSend][m]
		if ok1 && ok2 {
			// found potential mixed deadlock
			found := "Potential mixed deadlock:\n"
			found += "\tlock1: " + mostRecentAcquire[routineSend][m].tID + "\n"
			found += "\tlock2: " + mostRecentAcquire[routineRevc][m].tID

			logging.Result(found, logging.CRITICAL)
		}
	}

	for m := range lockSet[routineRevc] {
		_, ok1 := mostRecentAcquire[routineRevc][m]
		_, ok2 := mostRecentAcquire[routineSend][m]
		if ok1 && ok2 {
			// found potential mixed deadlock
			found := "Potential mixed deadlock:\n"
			found += "\tlock1: " + mostRecentAcquire[routineSend][m].tID + "\n"
			found += "\tlock2: " + mostRecentAcquire[routineRevc][m].tID

			logging.Result(found, logging.CRITICAL)
		}
	}
}

/*
func checkForMixedDeadlock2(routine int) {
	for m := range lockSet[routine] {
		// if the lock was not acquired by the routine, continue. Should not happen
		vc1, okS := mostRecentAcquire[routine][m]
		if !okS {
			continue
		}

		for routine2, acquire := range mostRecentAcquire {
			if routine == routine2 {
				continue
			}

			if vc2, ok := acquire[m]; ok {
				weakHappensBefore := GetHappensBefore(vc1, vc2)
				if weakHappensBefore != Concurrent {
					continue
				}

				// found potential mixed deadlock
				found := "Potential mixed deadlock:\n"
				found += "\tlock1: " + lockSet[routine][m] + "\n"
				found += "\tlock2: " + lockSet[routine2][m]

				logging.Result(found, logging.CRITICAL)
			}

		}
	}
}
*/
