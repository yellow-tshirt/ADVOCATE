package analysis

/*
 * Create a new relW and relR if needed
 * Args:
 *   index (int): The id of the atomic variable
 *   nRout (int): The number of routines in the trace
 */
func newRel(index int, nRout int) {
	if _, ok := relW[index]; !ok {
		relW[index] = NewVectorClock(nRout)
	}
	if _, ok := relR[index]; !ok {
		relR[index] = NewVectorClock(nRout)
	}
}

/*
 * Update and calculate the vector clocks given a lock operation
 * Args:
 *   routine (int): The routine id
 *   id (int): The id of the mutex
 *   vc (map[int]VectorClock): The current vector clocks
 *   wVc (map[int]VectorClock): The current weak vector clocks
 *   tID (string): The trace id of the lock operation
 *   tPost (int): The timestamp at the end of the event
 */
func Lock(routine int, id int, vc map[int]VectorClock, wVc map[int]VectorClock, tID string, tPost int) {
	if tPost == 0 {
		return
	}

	newRel(id, vc[routine].size)
	vc[routine] = vc[routine].Sync(relW[id])
	vc[routine] = vc[routine].Sync(relR[id])
	vc[routine] = vc[routine].Inc(routine)

	lockSetAddLock(routine, id, tID, wVc[routine])
}

/*
 * Update and calculate the vector clocks given a unlock operation
 * Args:
 *   routine (int): The routine id
 *   id (int): The id of the mutex
 *   vc (map[int]VectorClock): The current vector clocks
 */
func Unlock(routine int, id int, vc map[int]VectorClock, tPost int) {
	if tPost == 0 {
		return
	}

	newRel(id, vc[routine].size)
	relW[id] = vc[routine].Copy()
	relR[id] = vc[routine].Copy()
	vc[routine] = vc[routine].Inc(routine)

	lockSetRemoveLock(routine, id)
}

/*
 * Update and calculate the vector clocks given a rlock operation
 * Args:
 *   routine (int): The routine id
 *   id (int): The id of the mutex
 *   vc (map[int]VectorClock): The current vector clocks
 *   wVc (map[int]VectorClock): The current weak vector clocks
 *   tID (string): The trace id of the lock operation
 * Returns:
 *   (vectorClock): The new vector clock
 */
func RLock(routine int, id int, vc map[int]VectorClock, wVc map[int]VectorClock,
	tID string, tPost int) {

	if tPost != 0 {
		newRel(id, vc[routine].size)
		vc[routine] = vc[routine].Sync(relW[id])
		vc[routine] = vc[routine].Inc(routine)
	}

	// TODO: can we just add this to the lockSet?
	lockSetAddLock(routine, id, tID, wVc[routine])
}

/*
 * Update and calculate the vector clocks given a runlock operation
 * Args:
 *   routine (int): The routine id
 *   id (int): The id of the mutex
 *   vc (map[int]VectorClock): The current vector clocks
 *   tPost (int): The timestamp at the end of the event
 */
func RUnlock(routine int, id int, vc map[int]VectorClock, tPost int) {
	if tPost != 0 {
		newRel(id, vc[routine].size)
		relR[id] = relR[id].Sync(vc[routine])
		vc[routine] = vc[routine].Inc(routine)
	}

	lockSetRemoveLock(routine, id)
}
