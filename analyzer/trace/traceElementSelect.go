package trace

import (
	"errors"
	"strconv"
	"strings"
)

/*
 * traceElementSelect is a trace element for a select statement
 * Fields:
 *   routine (int): The routine id
 *   tpre (int): The timestamp at the start of the event
 *   tpost (int): The timestamp at the end of the event
 *   vpre (vectorClock): The vector clock at the start of the event
 *   vpost (vectorClock): The vector clock at the end of the event
 *   id (int): The id of the select statement
 *   cases ([]traceElementSelectCase): The cases of the select statement
 *   containsDefault (bool): Whether the select statement contains a default case
 *   chosenCase (traceElementSelectCase): The chosen case, nil if default case chosen
 *   chosenDefault (bool): if the default case was chosen
 *   pos (string): The position of the select statement in the code
 *   pre (*traceElementPre): The pre element of the select statement
 */
type traceElementSelect struct {
	routine         int
	tpre            int
	tpost           int
	vpre            vectorClock
	vpost           vectorClock
	id              int
	cases           []traceElementChannel
	containsDefault bool
	chosenDefault   bool
	pos             string
	pre             *traceElementPre
}

/*
 * Add a new select statement trace element
 * Args:
 *   routine (int): The routine id
 *   numberOfRoutines (int): The number of routines in the trace
 *   tpre (string): The timestamp at the start of the event
 *   tpost (string): The timestamp at the end of the event
 *   id (string): The id of the select statement
 *   cases (string): The cases of the select statement
 *   pos (string): The position of the select statement in the code
 */
func addTraceElementSelect(routine int, numberOfRoutines int, tpre string,
	tpost string, id string, cases string, pos string) error {
	tpre_int, err := strconv.Atoi(tpre)
	if err != nil {
		return errors.New("tpre is not an integer")
	}

	tpost_int, err := strconv.Atoi(tpost)
	if err != nil {
		return errors.New("tpost is not an integer")
	}

	id_int, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	elem := traceElementSelect{
		routine: routine,
		tpre:    tpre_int,
		tpost:   tpost_int,
		vpre:    newVectorClock(numberOfRoutines),
		vpost:   newVectorClock(numberOfRoutines),
		id:      id_int,
		pos:     pos,
	}

	cs := strings.Split(cases, "~")
	cases_list := make([]traceElementChannel, 0)
	containsDefault := false
	chosenDefault := false
	for _, c := range cs {
		if c == "d" {
			containsDefault = true
			break
		}
		if c == "D" {
			containsDefault = true
			chosenDefault = true
			break
		}

		// read channel operation
		case_list := strings.Split(c, ".")
		c_tpre, err := strconv.Atoi(case_list[1])
		if err != nil {
			return errors.New("c_tpre is not an integer")
		}
		c_tpost, err := strconv.Atoi(case_list[2])
		if err != nil {
			return errors.New("c_tpost is not an integer")
		}
		c_id, err := strconv.Atoi(case_list[3])
		if err != nil {
			return errors.New("c_id is not an integer")
		}
		var c_opC opChannel = send
		if case_list[4] == "R" {
			c_opC = recv
		} else if case_list[4] == "C" {
			panic("Close in select case list")
		}

		c_cl, err := strconv.ParseBool(case_list[5])
		if err != nil {
			return errors.New("c_cr is not a boolean")
		}

		c_oId, err := strconv.Atoi(case_list[6])
		if err != nil {
			return errors.New("c_oId is not an integer")
		}
		c_oSize, err := strconv.Atoi(case_list[7])
		if err != nil {
			return errors.New("c_oSize is not an integer")
		}

		elem := traceElementChannel{
			tpre:  c_tpre,
			tpost: c_tpost,
			id:    c_id,
			opC:   c_opC,
			cl:    c_cl,
			oId:   c_oId,
			qSize: c_oSize,
			sel:   &elem,
		}

		cases_list = append(cases_list, elem)
	}

	elem.containsDefault = containsDefault
	elem.chosenDefault = chosenDefault
	elem.cases = cases_list

	// create the pre event
	elem_pre := traceElementPre{
		elem:     &elem,
		elemType: Select,
	}
	elem.pre = &elem_pre

	err1 := addElementToTrace(&elem_pre)
	err2 := addElementToTrace(&elem)
	return errors.Join(err1, err2)
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (elem *traceElementSelect) getRoutine() int {
	return elem.routine
}

/*
 * Get the timestamp at the start of the event
 * Returns:
 *   int: The timestamp at the start of the event
 */
func (elem *traceElementSelect) getTpre() int {
	return elem.tpre
}

/*
 * Get the timestamp at the end of the event
 * Returns:
 *   int: The timestamp at the end of the event
 */
func (elem *traceElementSelect) getTpost() int {
	return elem.tpost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   int: The timer of the element
 */
func (elem *traceElementSelect) getTsort() float32 {
	if elem.tpost == 0 {
		return float32(elem.tpre) + 0.5
	}
	return float32(elem.tpost)
}

/*
 * Get the vector clock at the begin of the event
 * Returns:
 *   vectorClock: The vector clock at the begin of the event
 */
func (elem *traceElementSelect) getVpre() *vectorClock {
	return &elem.vpre
}

/*
 * Get the vector clock at the end of the event
 * Returns:
 *   vectorClock: The vector clock at the end of the event
 */
func (elem *traceElementSelect) getVpost() *vectorClock {
	return &elem.vpost
}

/*
 * Get the simple string representation of the element
 * Returns:
 *   string: The simple string representation of the element
 */
func (elem *traceElementSelect) toString() string {
	res := "S" + "," + strconv.Itoa(elem.tpre) + "," +
		strconv.Itoa(elem.tpost) + "," + strconv.Itoa(elem.id) + ","

	for i, c := range elem.cases {
		if i != 0 {
			res += "~"
		}
		res += c.toStringSep(".")
	}

	if elem.containsDefault {
		if elem.chosenDefault {
			res += ".D"
		} else {
			res += ".d"
		}
	}
	res += "," + elem.pos
	return res
}

/*
 * Update and calculate the vector clock of the element
 * Args:
 *   vc (vectorClock): The current vector clocks
 * TODO: implement
 */
func (elem *traceElementSelect) calculateVectorClock(vc *[]vectorClock) {
}
