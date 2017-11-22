package model

import (
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"strings"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
)

/** Constant and Variable Definitions */

const Duration60Days = 60 * 24 * time.Hour
const Duration90Days = 90 * 24 * time.Hour

const selectActivistBaseQuery string = `
SELECT
  chapter,
  email,
  facebook,
  id,
  location,
  name,
  phone
FROM activists
`

const selectActivistExtraBaseQuery string = `
SELECT

  chapter,
  email,
  facebook,
  a.id,
  location,
  a.name,
  phone,

  activist_level,
  core_staff,
  exclude_from_leaderboard,
  global_team_member,
  liberation_pledge,
  source,

  connector,
  contacted_date,
  core_training,
  eligable_senior_organizer,
  escalation,
  interested,
  meeting_date,

  MIN(e.date) AS first_event,
  MAX(e.date) AS last_event,
  COUNT(e.id) as total_events
FROM activists a

LEFT JOIN event_attendance ea
  ON ea.activist_id = a.id

LEFT JOIN events e
  ON ea.event_id = e.id
`

const DescOrder int = 2
const AscOrder int = 1

/** Type Definitions */

type Activist struct {
	Chapter  string         `db:"chapter"`
	Email    string         `db:"email"`
	Facebook string         `db:"facebook"`
	Hidden   bool           `db:"hidden"`
	ID       int            `db:"id"`
	Location sql.NullString `db:"location"`
	Name     string         `db:"name"`
	Phone    string         `db:"phone"`
}

type ActivistEventData struct {
	FirstEvent  *time.Time `db:"first_event"`
	LastEvent   *time.Time `db:"last_event"`
	TotalEvents int        `db:"total_events"`
	Status      string
}

type ActivistMembershipData struct {
	ActivistLevel          string `db:"activist_level"`
	CoreStaff              bool   `db:"core_staff"`
	ExcludeFromLeaderboard bool   `db:"exclude_from_leaderboard"`
	GlobalTeamMember       bool   `db:"global_team_member"`
	LiberationPledge       bool   `db:"liberation_pledge"`
	Source                 string `db:"source"`
}

type ActivistConnectionData struct {
	Connector               string `db:"connector"`
	ContactedDate           string `db:"contacted_date"`
	CoreTraining            bool   `db:"core_training"`
	EligableSeniorConnector bool   `db:"eligable_senior_organizer"`
	Escalation              string `db:"escalation"`
	Interested              string `db:"interested"`
	MeetingDate             string `db:"meeting_date"`
}

type ActivistExtra struct {
	Activist
	ActivistEventData
	ActivistMembershipData
	ActivistConnectionData
}

type ActivistJSON struct {
	Chapter  string `json:"chapter"`
	Email    string `json:"email"`
	Facebook string `json:"facebook"`
	ID       int    `json:"id"`
	Location string `json:"location"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`

	FirstEvent  string `json:"first_event"`
	LastEvent   string `json:"last_event"`
	TotalEvents int    `json:"total_events"`
	Status      string `json:"status"`

	ActivistLevel          string `json:"activist_level"`
	CoreStaff              bool   `json:"core_staff"`
	ExcludeFromLeaderboard bool   `json:"exclude_from_leaderboard"`
	GlobalTeamMember       bool   `json:"global_team_member"`
	LiberationPledge       bool   `json:"liberation_pledge"`
	Source                 string `json:"source"`

	Connector               string `json:"connector"`
	ContactedDate           string `json:"contacted_date"`
	CoreTraining            bool   `json:"core_training"`
	EligableSeniorConnector bool   `json:"eligable_senior_organizer"`
	Escalation              string `json:"escalation"`
	Interested              string `json:"interested"`
	MeetingDate             string `json:"meeting_date"`
}

type GetActivistOptions struct {
	ID     int
	Hidden bool
}

type ActivistRangeOptionsJSON struct {
	Name              string `json:"name"`
	Limit             int    `json:"limit"`
	Order             int    `json:"order"`
	LastEventDateFrom string `json:"last_event_date_from"`
	LastEventDateTo   string `json:"last_event_date_to"`
}

/** Functions and Methods */

func GetActivistsJSON(db *sqlx.DB, options GetActivistOptions) ([]ActivistJSON, error) {
	if options.ID != 0 {
		return nil, errors.New("GetActivistsJSON: Cannot include ID in options")
	}
	return getActivistsJSON(db, options)
}

func GetActivistJSON(db *sqlx.DB, options GetActivistOptions) (ActivistJSON, error) {
	if options.ID == 0 {
		return ActivistJSON{}, errors.New("GetActivistJSON: Must include ID in options")
	}

	activists, err := getActivistsJSON(db, options)
	if err != nil {
		return ActivistJSON{}, err
	} else if len(activists) == 0 {
		return ActivistJSON{}, errors.New("Could not find any activists")
	} else if len(activists) > 1 {
		return ActivistJSON{}, errors.New("Found too many activists")
	}
	return activists[0], nil
}

func getActivistsJSON(db *sqlx.DB, options GetActivistOptions) ([]ActivistJSON, error) {
	activists, err := GetActivistsExtra(db, options)
	if err != nil {
		return nil, err
	}
	return buildActivistJSONArray(activists), nil
}

func GetActivistRangeJSON(db *sqlx.DB, activistOptions ActivistRangeOptionsJSON) ([]ActivistJSON, error) {
	// Check that order matches one of the defined order constants
	if activistOptions.Order != DescOrder && activistOptions.Order != AscOrder {
		return nil, errors.New("User Range order must be ascending or descending")
	}
	activists, err := getActivistRange(db, activistOptions)
	if err != nil {
		return nil, err
	}
	return buildActivistJSONArray(activists), nil
}

func buildActivistJSONArray(activists []ActivistExtra) []ActivistJSON {
	var activistsJSON []ActivistJSON

	for _, a := range activists {
		firstEvent := ""
		if a.ActivistEventData.FirstEvent != nil {
			firstEvent = a.ActivistEventData.FirstEvent.Format(EventDateLayout)
		}
		lastEvent := ""
		if a.ActivistEventData.LastEvent != nil {
			lastEvent = a.ActivistEventData.LastEvent.Format(EventDateLayout)
		}
		location := ""
		if a.Activist.Location.Valid {
			location = a.Activist.Location.String
		}

		activistsJSON = append(activistsJSON, ActivistJSON{
			Chapter:  a.Chapter,
			Email:    a.Email,
			Facebook: a.Facebook,
			ID:       a.ID,
			Location: location,
			Name:     a.Name,
			Phone:    a.Phone,

			FirstEvent:  firstEvent,
			LastEvent:   lastEvent,
			Status:      a.Status,
			TotalEvents: a.TotalEvents,

			ActivistLevel:          a.ActivistLevel,
			CoreStaff:              a.CoreStaff,
			ExcludeFromLeaderboard: a.ExcludeFromLeaderboard,
			GlobalTeamMember:       a.GlobalTeamMember,
			LiberationPledge:       a.LiberationPledge,
			Source:                 a.Source,

			Connector:               a.Connector,
			ContactedDate:           a.ContactedDate,
			CoreTraining:            a.CoreTraining,
			EligableSeniorConnector: a.EligableSeniorConnector,
			Escalation:              a.Escalation,
			Interested:              a.Interested,
			MeetingDate:             a.MeetingDate,
		})
	}

	return activistsJSON
}

func GetActivist(db *sqlx.DB, name string) (Activist, error) {
	activists, err := getActivists(db, name)
	if err != nil {
		return Activist{}, err
	} else if len(activists) == 0 {
		return Activist{}, errors.New("Could not find any activists")
	} else if len(activists) > 1 {
		return Activist{}, errors.New("Found too many activists")
	}
	return activists[0], nil
}

func GetActivists(db *sqlx.DB) ([]Activist, error) {
	return getActivists(db, "")
}

func getActivists(db *sqlx.DB, name string) ([]Activist, error) {
	var queryArgs []interface{}
	query := selectActivistBaseQuery

	if name != "" {
		query += " WHERE name = ? "
		queryArgs = append(queryArgs, name)
	}

	query += " ORDER BY name "

	var activists []Activist
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get activists for %s", name)
	}

	return activists, nil
}

func GetActivistsExtra(db *sqlx.DB, options GetActivistOptions) ([]ActivistExtra, error) {
	query := selectActivistExtraBaseQuery

	var queryArgs []interface{}

	if options.ID != 0 {
		// retrieve specific activist rather than all activists
		query += " WHERE a.id = ? "
		queryArgs = append(queryArgs, options.ID)
	} else {
		// Only check filter by hidden if the activistID isn't
		// supplied.
		if options.Hidden == true {
			query += " WHERE a.hidden = true "
		} else {
			query += " WHERE a.hidden = false "
		}
	}

	query += " GROUP BY a.id "

	var activists []ActivistExtra
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get activists extra for uid %d", options.ID)
	}

	for i := 0; i < len(activists); i++ {
		a := activists[i]
		activists[i].Status = getStatus(a.FirstEvent, a.LastEvent, a.TotalEvents)
	}

	return activists, nil
}

// TODO Make sure you only fetch non-hidden members
// THAT is not currently the case
func getActivistRange(db *sqlx.DB, activistOptions ActivistRangeOptionsJSON) ([]ActivistExtra, error) {
	query := selectActivistExtraBaseQuery
	name := activistOptions.Name
	order := activistOptions.Order
	limit := activistOptions.Limit
	var queryArgs []interface{}

	query += " WHERE a.hidden = false "

	if name != "" {
		if order == DescOrder {
			query += " AND a.name < ? "
		} else {
			query += " AND a.name > ? "
		}
		queryArgs = append(queryArgs, name)
	}

	query += " GROUP BY a.name "

	havingClause := []string{}
	if activistOptions.LastEventDateFrom != "" {
		havingClause = append(havingClause, "last_event >= ?")
		queryArgs = append(queryArgs, activistOptions.LastEventDateFrom)
	}
	if activistOptions.LastEventDateTo != "" {
		havingClause = append(havingClause, "last_event <= ?")
		queryArgs = append(queryArgs, activistOptions.LastEventDateTo)
	}

	if len(havingClause) != 0 {
		query += " HAVING " + strings.Join(havingClause, " AND ")
	}

	query += " ORDER BY a.name "
	if order == DescOrder {
		query += "desc "
	}

	if limit > 0 {
		query += " LIMIT ? "
		queryArgs = append(queryArgs, limit)
	}

	var activists []ActivistExtra
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve %d users before/after %s", limit, name)
	}

	for i := 0; i < len(activists); i++ {
		a := activists[i]
		activists[i].Status = getStatus(a.FirstEvent, a.LastEvent, a.TotalEvents)
	}

	return activists, nil
}

func (a Activist) GetActivistEventData(db *sqlx.DB) (ActivistEventData, error) {
	query := `
SELECT
  MIN(e.date) AS first_event,
  MAX(e.date) AS last_event,
  COUNT(*) as total_events
FROM events e
JOIN event_attendance
  ON event_attendance.event_id = e.id
WHERE
  event_attendance.activist_id = ?
`
	var data ActivistEventData
	if err := db.Get(&data, query, a.ID); err != nil {
		return ActivistEventData{}, errors.Wrap(err, "failed to get activist event data")
	}
	return data, nil
}

func GetOrCreateActivist(db *sqlx.DB, name string) (Activist, error) {
	activist, err := GetActivist(db, name)
	if err == nil {
		// We got a valid activist, return them.
		return activist, nil
	}

	// There was an error, so try inserting the activist first.
	// Wrap in transaction to avoid issue where a new activist
	// is inserted successfully, but we are unable to retrieve
	// the new activist, which will leave database in inconsistent state

	tx, err := db.Beginx()
	if err != nil {
		return Activist{}, errors.Wrap(err, "Failed to create transaction")
	}

	_, err = tx.Exec("INSERT INTO activists (name) VALUES (?)", name)
	if err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to insert activist %s", name)
	}

	query := selectActivistBaseQuery + " WHERE name = ? "

	var newActivist Activist
	err = tx.Get(&newActivist, query, name)

	if err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to get new activist %s", name)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to commit activist %s", name)
	}

	return newActivist, nil
}

func CreateActivist(db *sqlx.DB, activist ActivistExtra) (int, error) {
	if activist.ID != 0 {
		return 0, errors.New("Activist ID must be 0")
	}
	if activist.Name == "" {
		return 0, errors.New("Name cannot be empty")
	}

	result, err := db.NamedExec(`
INSERT INTO activists (

  chapter,
  email,
  facebook,
  location,
  name,
  phone,

  activist_level,
  core_staff,
  exclude_from_leaderboard,
  global_team_member,
  liberation_pledge,
  source,

  connector,
  contacted_date,
  core_training,
  eligable_senior_organizer,
  escalation,
  interested,
  meeting_date

) VALUES (

  :chapter,
  :email,
  :facebook,
  :location,
  :name,
  :phone,

  :activist_level,
  :core_staff,
  :exclude_from_leaderboard,
  :global_team_member,
  :liberation_pledge,
  :source,

  :connector,
  :contacted_date,
  :core_training,
  :eligable_senior_organizer,
  :escalation,
  :interested,
  :meeting_date

)`, activist)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not create activist: %s", activist.Name)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrapf(err, "Could not get LastInsertId for %s", activist.Name)
	}
	return int(id), nil
}

func UpdateActivistData(db *sqlx.DB, activist ActivistExtra) (int, error) {
	if activist.ID == 0 {
		return 0, errors.New("activist ID cannot be 0")
	}
	if activist.Name == "" {
		return 0, errors.New("Name cannot be empty")
	}

	_, err := db.NamedExec(`UPDATE activists
SET

  chapter = :chapter,
  email = :email,
  facebook = :facebook,
  location = :location,
  name = :name,
  phone = :phone,

  activist_level = :activist_level,
  core_staff = :core_staff,
  exclude_from_leaderboard = :exclude_from_leaderboard,
  global_team_member = :global_team_member,
  liberation_pledge = :liberation_pledge,
  source = :source,

  connector = :connector,
  contacted_date = :contacted_date,
  core_training = :core_training,
  eligable_senior_organizer = :eligable_senior_organizer,
  escalation = :escalation,
  interested = :interested,
  meeting_date = :meeting_date

WHERE
  id = :id`, activist)

	if err != nil {
		return 0, errors.Wrap(err, "failed to update activist data")
	}
	return activist.ID, nil
}

func HideActivist(db *sqlx.DB, activistID int) error {
	if activistID == 0 {
		return errors.New("HideActivist: activistID cannot be 0")
	}
	var activistCount int
	err := db.Get(&activistCount, `SELECT count(*) FROM activists WHERE id = ?`, activistID)
	if err != nil {
		return errors.Wrap(err, "failed to get activist count")
	}
	if activistCount == 0 {
		return errors.Errorf("Activist with id %d does not exist", activistID)
	}

	_, err = db.Exec(`UPDATE activists SET hidden = true WHERE id = ?`, activistID)
	if err != nil {
		return errors.Wrapf(err, "failed to update activist %d", activistID)
	}
	return nil
}

// Merge activistID into targetActivistID.
//  - The original activist is hidden
//  - All of the original activist's event attendance is updated to be the target activist.
func MergeActivist(db *sqlx.DB, originalActivistID, targetActivistID int) error {
	if originalActivistID == 0 {
		return errors.New("originalActivistID cannot be 0")
	}
	if targetActivistID == 0 {
		return errors.New("targetActivistID cannot be 0")
	}
	if originalActivistID == targetActivistID {
		return errors.New("originalActivist and targetActivist cannot be the same")
	}

	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "could not create transaction")
	}

	_, err = tx.Exec(`UPDATE activists SET hidden = true, name = concat(name,' ', id) WHERE id = ?`, originalActivistID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "failed to hide original activist %d", originalActivistID)
	}

	err = updateMergedActivistData(tx, originalActivistID, targetActivistID, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = updateMergedActivistData(tx, originalActivistID, targetActivistID, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrapf(err,
			"failed to commit merge activist transaction. original activist id: %d, target activist id: %d",
			originalActivistID, targetActivistID)
	}

	return nil
}

func updateMergedActivistData(tx *sqlx.Tx, originalActivistID int, targetActivistID int, originalActivistOnly bool) error {
	baseQuery := `
SELECT event_id
FROM event_attendance ea
WHERE
  activist_id = ?
  AND `

	subquery := `
EXISTS(
  SELECT ea2.event_id
  FROM event_attendance ea2
  WHERE ea2.activist_id = ?
    AND ea2.event_id = ea.event_id)`

	var eventQuery string
	if originalActivistOnly {
		eventQuery = baseQuery + " NOT " + subquery
	} else {
		eventQuery = baseQuery + subquery
	}

	var eventIDs []int
	err := tx.Select(&eventIDs, eventQuery, originalActivistID, targetActivistID)
	if err != nil {
		return errors.Wrapf(err,
			"failed to get original activist's events: %d, originalActivistOnly: %v",
			originalActivistID, originalActivistOnly)
	}

	// There's nothing to do if there are no events.
	if len(eventIDs) == 0 {
		return nil
	}

	var eaQuery string
	var eaArgs []interface{}
	if originalActivistOnly {
		eaQuery, eaArgs, err = sqlx.In(`
UPDATE event_attendance
SET activist_id = ?
WHERE
  activist_id = ?
  AND event_id IN (?)`,
			targetActivistID, originalActivistID, eventIDs)
	} else {
		eaQuery, eaArgs, err = sqlx.In(`
DELETE FROM event_attendance
WHERE
  activist_id = ?
  AND event_id IN (?)`, originalActivistID, eventIDs)
	}
	if err != nil {
		return errors.Wrapf(err, "could not create sqlx.IN query. originalActivistOnly: %v",
			originalActivistOnly)
	}
	eaQuery = tx.Rebind(eaQuery)
	_, err = tx.Exec(eaQuery, eaArgs...)
	if err != nil {
		return errors.Wrapf(err, "could not update event attendance for activist: %d",
			originalActivistID)
	}
	err = insertMergedActivistAttendance(tx, originalActivistID, targetActivistID, eventIDs, originalActivistOnly)
	if err != nil {
		return err
	}
	return nil
}

func insertMergedActivistAttendance(tx *sqlx.Tx, originalActivistID int, targetActivistID int, eventIDs []int, replacedWithTargetActivist bool) error {
	if len(eventIDs) == 0 {
		return nil
	}

	query := `INSERT INTO merged_activist_attendance (original_activist_id, target_activist_id, event_id, replaced_with_target_activist) VALUES`

	var queryValues []string
	var queryArgs []interface{}
	for _, eventID := range eventIDs {
		queryValues = append(queryValues, " (?, ?, ?, ?) ")
		queryArgs = append(queryArgs, originalActivistID, targetActivistID, eventID, replacedWithTargetActivist)
	}
	query += strings.Join(queryValues, ",")
	_, err := tx.Exec(query, queryArgs...)

	return errors.Wrapf(err, "could not insert merged_activist_attendance for originalActivistID: %d, targetActivistID: %d",
		originalActivistID, targetActivistID)
}

func GetAutocompleteNames(db *sqlx.DB) []string {
	type Name struct {
		Name string `db:"name"`
	}
	names := []Name{}
	// Order the activists by the last even they've been to.
	err := db.Select(&names, `
SELECT a.name FROM activists a
LEFT OUTER JOIN event_attendance ea ON a.id = ea.activist_id
LEFT OUTER JOIN events e ON e.id = ea.event_id
WHERE a.hidden = 0
GROUP BY a.name
ORDER BY MAX(e.date) DESC`)
	if err != nil {
		// TODO: return error
		panic(err)
	}

	ret := []string{}
	for _, n := range names {
		ret = append(ret, n.Name)
	}
	return ret
}

func CleanActivistData(body io.Reader) (ActivistExtra, error) {
	var activistJSON ActivistJSON
	err := json.NewDecoder(body).Decode(&activistJSON)
	if err != nil {
		return ActivistExtra{}, err
	}

	// Check if name field contains dangerous input
	if err := checkForDangerousChars(activistJSON.Name); err != nil {
		return ActivistExtra{}, err
	}

	valid := true
	if activistJSON.Location == "" {
		// No location specified so insert null value into database
		valid = false
	}

	activistExtra := ActivistExtra{
		Activist: Activist{
			Chapter:  strings.TrimSpace(activistJSON.Chapter),
			Email:    strings.TrimSpace(activistJSON.Email),
			Facebook: strings.TrimSpace(activistJSON.Facebook),
			ID:       activistJSON.ID,
			Location: sql.NullString{String: strings.TrimSpace(activistJSON.Location), Valid: valid},
			Name:     strings.TrimSpace(activistJSON.Name),
			Phone:    strings.TrimSpace(activistJSON.Phone),
		},
		ActivistMembershipData: ActivistMembershipData{
			ActivistLevel:          strings.TrimSpace(activistJSON.ActivistLevel),
			CoreStaff:              activistJSON.CoreStaff,
			ExcludeFromLeaderboard: activistJSON.ExcludeFromLeaderboard,
			GlobalTeamMember:       activistJSON.GlobalTeamMember,
			LiberationPledge:       activistJSON.LiberationPledge,
			Source:                 strings.TrimSpace(activistJSON.Source),
		},
		ActivistConnectionData: ActivistConnectionData{
			Connector:               strings.TrimSpace(activistJSON.Connector),
			ContactedDate:           strings.TrimSpace(activistJSON.ContactedDate),
			CoreTraining:            activistJSON.CoreTraining,
			EligableSeniorConnector: activistJSON.EligableSeniorConnector,
			Escalation:              strings.TrimSpace(activistJSON.Escalation),
			Interested:              strings.TrimSpace(activistJSON.Interested),
			MeetingDate:             strings.TrimSpace(activistJSON.MeetingDate),
		},
	}

	if err := validateActivist(activistExtra); err != nil {
		return ActivistExtra{}, err
	}

	return activistExtra, nil

}

var validActivistLevels = map[string]struct{}{
	"activist":         struct{}{},
	"not_local":        struct{}{},
	"organizer":        struct{}{},
	"hiatus":           struct{}{},
	"prospect":         struct{}{},
	"senior_organizer": struct{}{},
	"none":             struct{}{},
}

func validateActivist(a ActivistExtra) error {
	if _, ok := validActivistLevels[a.ActivistLevel]; !ok {
		return errors.New("ActivistLevel must be one of: activist, not_local, " +
			"organizer, hiatus, prospect, senior_organizer, none")
	}
	return nil
}

func GetActivistRangeOptions(body io.Reader) (ActivistRangeOptionsJSON, error) {
	var activistOptions ActivistRangeOptionsJSON
	err := json.NewDecoder(body).Decode(&activistOptions)
	if err != nil {
		return ActivistRangeOptionsJSON{}, err
	}
	return activistOptions, nil
}

// Returns one of the following statuses:
//  - Current
//  - New
//  - Former
//  - No attendance
// Must be kept in sync with the list in frontend/ActivistList.vue
func getStatus(firstEvent *time.Time, lastEvent *time.Time, totalEvents int) string {
	if firstEvent == nil || lastEvent == nil {
		return "No attendance"
	}

	if time.Since(*lastEvent) > Duration60Days {
		return "Former"
	}
	if time.Since(*firstEvent) < Duration90Days && totalEvents < 5 {
		return "New"
	}
	return "Current"
}
