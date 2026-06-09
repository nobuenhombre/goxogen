// pgxdb.DBQuery is the common interface for database operations that can be used with
// types from schema '{{ schema .Schema }} '.

func CheckErr(err error) {
    if err != nil {
        panic(err)
    }
}

// pg_catalog types
type Abstime struct{}
type Aclitem struct{}
type Anyarray struct{}
type Anyelement struct{}
type Anyenum struct{}
type Anynonarray struct{}
type Anyrange struct{}
type Box struct{}
type Cidr struct{}
type Cid struct{}
type Circle struct{}
type Cstring struct{}
type Daterange struct{}
type Date struct{}
type EventTrigger struct{}
type FdwHandler struct{}
type Gtsvector struct{}
type Inet struct{}
type Int2vector struct{}
type Int4range struct{}
type Int8range struct{}
type Internal struct{}
type Int struct{}
type LanguageHandler struct{}
type Line struct{}
type Lseg struct{}
type Macaddr struct{}
type Name struct{}
type Numrange struct{}
type Oid struct{}
type Oidvector struct{}
type Opaque struct{}
type Path struct{}
type PgDdlCommand struct{}
type PgLsn struct{}
type PgNodeTree struct{}
type Point struct{}
type Polygon struct{}
type Record struct{}
type Refcursor struct{}
type Regclass struct{}
type Regconfig struct{}
type Regdictionary struct{}
type Regnamespace struct{}
type Regoperator struct{}
type Regoper struct{}
type Regprocedure struct{}
type Regproc struct{}
type Regrole struct{}
type Regtype struct{}
type Reltime struct{}
type Smgr struct{}
type SQLIdentifier struct{}
type Tid struct{}
type TimeStamp struct{}
type Tinterval struct{}
type Trigger struct{}
type TsmHandler struct{}
type Tsquery struct{}
type Tsrange struct{}
type Tstzrange struct{}
type Tsvector struct{}
type TxidSnapshot struct{}
type Unknown struct{}
type UUID struct{}
type Void struct{}
type Xid struct{}
type XML struct{}

// postgis types
type GeometryPoint struct {
    Lng float64 `json:"lng"`
    Lat float64 `json:"lat"`
}

func (g GeometryPoint) Value() (driver.Value, error) {
    return fmt.Sprintf("POINT (%f %f)", g.Lng, g.Lat), nil
}

func (g *GeometryPoint) Scan(value interface{}) error {
    if value == nil {
        g.Lng = 0
        g.Lat = 0
        return nil
    }

    var data string
    switch v := value.(type) {
    case []byte:
        data = string(v)
    case string:
        data = v
    default:
        return fmt.Errorf("GeometryPoint: unsupported type: %T", value)
    }

    return g.parseWKT(data)
}

func (g *GeometryPoint) parseWKT(data string) error {
    // WKT: POINT (x y) or EWKT: SRID=N;POINT (x y)
    re := regexp.MustCompile(`(?:SRID=\d+;)?\s*POINT\s*\(\s*([\d\.\-eE]+)\s+([\d\.\-eE]+)\s*\)`)
    matches := re.FindStringSubmatch(data)
    if matches == nil {
        return fmt.Errorf("GeometryPoint: cannot parse WKT: %s", data)
    }

    lng, err := strconv.ParseFloat(matches[1], 64)
    if err != nil {
        return fmt.Errorf("GeometryPoint: invalid lng %q: %w", matches[1], err)
    }

    lat, err := strconv.ParseFloat(matches[2], 64)
    if err != nil {
        return fmt.Errorf("GeometryPoint: invalid lat %q: %w", matches[2], err)
    }

    g.Lng = lng
    g.Lat = lat
    return nil
}

// Geometry stores a PostGIS geometry as WKT string (works with any geometry type:
// Point, MultiPolygon, LineString, Polygon, etc.).
type Geometry struct {
    WKT string `json:"wkt"`
}

func (g Geometry) Value() (driver.Value, error) {
    return g.WKT, nil
}

func (g *Geometry) Scan(value interface{}) error {
    if value == nil {
        g.WKT = ""
        return nil
    }

    switch v := value.(type) {
    case []byte:
        g.WKT = string(v)
    case string:
        g.WKT = v
    default:
        return fmt.Errorf("Geometry: unsupported type: %T", value)
    }

    return nil
}

// information_schema types
type CharacterData struct{}
type PgAttribute struct{}
type PgType struct{}
type YesOrNo struct{}
type CardinalNumber struct{}

type JSON struct {
    Data interface{}
}

func (j JSON) Value() (driver.Value, error) {
    return json.Marshal(j.Data)
}

func (j *JSON) Scan(value interface{}) error {
    if value == nil {
        j.Data = nil

        return nil
    }

    var data []byte
    switch v := value.(type) {
    case []byte:
        data = v
    case string:
        data = []byte(v)
    default:
        return fmt.Errorf("unsupported type: %T", value)
    }

    return json.Unmarshal(data, &j.Data)
}

// Вспомогательные методы для удобства
func (j *JSON) AsObject() (map[string]interface{}, bool) {
    obj, ok := j.Data.(map[string]interface{})

    return obj, ok
}

func (j *JSON) AsArray() ([]map[string]interface{}, bool) {
    arr, ok := j.Data.([]interface{})
    if !ok {
        return nil, false
    }

    // Конвертируем []interface{} в []map[string]interface{}
    result := make([]map[string]interface{}, len(arr))
    for i, item := range arr {
        if m, ok := item.(map[string]interface{}); ok {
            result[i] = m
        } else {
            return nil, false
        }
    }

	return result, true
}

type Jsonb struct {
    Data interface{}
}

func (j Jsonb) Value() (driver.Value, error) {
    return json.Marshal(j.Data)
}

func (j *JSON) UnmarshalJSON(data []byte) error {
    return json.Unmarshal(data, &j.Data)
}

func (j *Jsonb) Scan(value interface{}) error {
    if value == nil {
        j.Data = nil

        return nil
    }

    var data []byte
    switch v := value.(type) {
    case []byte:
        data = v
    case string:
        data = []byte(v)
    default:
        return fmt.Errorf("unsupported type: %T", value)
    }

    return json.Unmarshal(data, &j.Data)
}

// Вспомогательные методы для удобства
func (j *Jsonb) AsObject() (map[string]interface{}, bool) {
    obj, ok := j.Data.(map[string]interface{})

    return obj, ok
}

func (j *Jsonb) AsArray() ([]map[string]interface{}, bool) {
    arr, ok := j.Data.([]interface{})
    if !ok {
        return nil, false
    }

    // Конвертируем []interface{} в []map[string]interface{}
    result := make([]map[string]interface{}, len(arr))
    for i, item := range arr {
        if m, ok := item.(map[string]interface{}); ok {
            result[i] = m
        } else {
            return nil, false
        }
    }

    return result, true
}
