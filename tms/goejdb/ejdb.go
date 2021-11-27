package goejdb

// #cgo LDFLAGS: -lejdb
// #include <ejdb/ejdb.h>
import "C"

import (
    "errors"
    "fmt"
    "unsafe"
)

// Database open modes
const (
    // Open as a reader.
    JBOREADER = C.JBOREADER
    // Open as a writer.
    JBOWRITER = C.JBOWRITER
    // Create if db file not exists.
    JBOCREAT  = C.JBOCREAT
    // Truncate db on open.
    JBOTRUNC  = C.JBOTRUNC
    // Open without locking.
    JBONOLCK  = C.JBONOLCK
    // Lock without blocking.
    JBOLCKNB  = C.JBOLCKNB
    // Synchronize every transaction.
    JBOTSYNC  = C.JBOTSYNC
)

// Error codes
const (
    // Invalid collection name.
    JBEINVALIDCOLNAME = C.JBEINVALIDCOLNAME
    // Invalid bson object.
    JBEINVALIDBSON = C.JBEINVALIDBSON
    // Invalid bson object id.
    JBEINVALIDBSONPK = C.JBEINVALIDBSONPK
    // Invalid query control field starting with '$'.
    JBEQINVALIDQCONTROL = C.JBEQINVALIDQCONTROL
    // $strand, $stror, $in, $nin, $bt keys requires not empty array value.
    JBEQINOPNOTARRAY = C.JBEQINOPNOTARRAY
    // Inconsistent database metadata.
    JBEMETANVALID = C.JBEMETANVALID
    // Invalid field path value.
    JBEFPATHINVALID = C.JBEFPATHINVALID
    // Invalid query regexp value.
    JBEQINVALIDQRX = C.JBEQINVALIDQRX
    // Result set sorting error.
    JBEQRSSORTING = C.JBEQRSSORTING
    // Query generic error.
    JBEQERROR = C.JBEQERROR
    // Updating failed.
    JBEQUPDFAILED = C.JBEQUPDFAILED
    // Only one $elemMatch allowed in the fieldpath.
    JBEQONEEMATCH = C.JBEQONEEMATCH
    // $fields hint cannot mix include and exclude fields
    JBEQINCEXCL = C.JBEQINCEXCL
    // action key in $do block can only be one of: $join
    JBEQACTKEY = C.JBEQACTKEY
    // Exceeded the maximum number of collections per database
    JBEMAXNUMCOLS = C.JBEMAXNUMCOLS
)

const maxslice = 0x7FFFFFFF

// An EJDB database
type Ejdb struct {
    ptr *[0]byte
}

type EjdbError struct {
    // Error code returned by EJDB
    ErrorCode int
    error
}

// EJDB collection tuning options
type EjCollOpts struct {
    // Large collection. It can be larger than 2GB. Default false
    Large         bool
    // Collection records will be compressed with DEFLATE compression. Default: false
    Compressed    bool
    // Expected records number in the collection. Default: 128K
    Records       int
    // Maximum number of cached records. Default: 0
    CachedRecords int
}

func new_ejdb() *Ejdb {
    ejdb := new(Ejdb)
    ejdb.ptr = (*[0]byte)(unsafe.Pointer(C.ejdbnew()))
    if ejdb.ptr == nil {
        return nil
    }
    return ejdb
}

// Returns EJDB library version string. Eg: "1.1.13"
func Version() string {
    cs := C.ejdbversion()
    return C.GoString(cs)
}

// Return true if passed `oid` string cat be converted to valid 12 bit BSON object identifier (OID).
func IsValidOidStr(oid string) bool {
    c_oid := C.CString(oid)
    res := C.ejdbisvalidoidstr(c_oid)
    C.free(unsafe.Pointer(c_oid))

    return bool(res)
}

// Returns a new open EJDB database.
// path is the path to the database file.
// options specify the open mode bitmask flags.
func Open(path string, options int) (*Ejdb, *EjdbError) {
    ejdb := new_ejdb()
    if ejdb != nil {
        c_path := C.CString(path)
        defer C.free(unsafe.Pointer(c_path))
        C.ejdbopen((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_path, C.int(options))
    }

    return ejdb, ejdb.check_error()
}

func (ejdb *Ejdb) check_error() *EjdbError {
    ecode := C.ejdbecode((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)))
    if ecode == 0 {
        return nil
    }
    c_msg := C.ejdberrmsg(ecode)
    msg := C.GoString(c_msg)
    return &EjdbError{int(ecode), errors.New(fmt.Sprintf("EJDB error: %v", msg))}
}

// Return true if database is in open state, false otherwise
func (ejdb *Ejdb) IsOpen() bool {
    ret := C.ejdbisopen((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)))
    return bool(ret)
}

// Delete database object. If the database is not closed, it is closed implicitly.
// Note that the deleted object and its derivatives can not be used anymore
func (ejdb *Ejdb) Del() {
    C.ejdbdel((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)))
}

// Close a table database object. If a writer opens a database but does not close it appropriately, the database will be broken.
// If successful return true, otherwise return false.
func (ejdb *Ejdb) Close() *EjdbError {
    C.ejdbclose((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)))
    return ejdb.check_error()
}

// Retrieve collection handle for collection specified `collname`.
// If collection with specified name does't exists it will return nil.
func (ejdb *Ejdb) GetColl(colname string) (*EjColl, *EjdbError) {
    c_colname := C.CString(colname)
    defer C.free(unsafe.Pointer(c_colname))

    ejcoll := new(EjColl)
    ejcoll.ejdb = ejdb
    ejcoll.ptr = (*[0]byte)(unsafe.Pointer(C.ejdbgetcoll((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_colname)))

    if ejcoll.ptr == nil {
        currentError := ejdb.check_error()
        if currentError == nil {
            currentError = &EjdbError{9000, errors.New("No such collection")}
        }
        return nil, currentError
    }
    return ejcoll, nil
}

// Return a slice containing shallow copies of all collection handles (EjColl) currently open.
func (ejdb *Ejdb) GetColls() ([]*EjColl, *EjdbError) {
    ret := make([]*EjColl, 0)
    lst := C.ejdbgetcolls((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)))
    if lst == nil {
        return ret, ejdb.check_error()
    }

    for i := int(lst.start); i < int(lst.start)+int(lst.num); i++ {
        ptr := uintptr(unsafe.Pointer(lst.array)) + unsafe.Sizeof(C.TCLISTDATUM{})*uintptr(i)
        datum := (*C.TCLISTDATUM)(unsafe.Pointer(ptr))
        datum_ptr := unsafe.Pointer(datum.ptr)
        ret = append(ret, &EjColl{(*[0]byte)(datum_ptr), ejdb})
    }
    return ret, nil
}

// Same as GetColl() but automatically creates new collection if it doesn't exists.
func (ejdb *Ejdb) CreateColl(colname string, opts *EjCollOpts) (*EjColl, *EjdbError) {
    c_colname := C.CString(colname)
    defer C.free(unsafe.Pointer(c_colname))

    ret := new(EjColl)
    ret.ejdb = ejdb

    if opts != nil {
        var c_opts C.EJCOLLOPTS
        c_opts.large = C._Bool(opts.Large)
        c_opts.compressed = C._Bool(opts.Large)
        c_opts.records = C.int64_t(opts.Records)
        c_opts.cachedrecords = C.int(opts.CachedRecords)
        ret.ptr = (*[0]byte)(unsafe.Pointer(C.ejdbcreatecoll((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_colname, &c_opts)))
    } else {
        ret.ptr = (*[0]byte)(unsafe.Pointer(C.ejdbcreatecoll((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_colname, nil)))
    }

    if ret.ptr != nil {
        return ret, nil
    }
    return nil, ejdb.check_error()
}

// Removes collections specified by `colname`.
// If `unlinkfile` is true the collection db file and all of its index files will be removed.
// If removal was successful return true, otherwise return false.
func (ejdb *Ejdb) RmColl(colname string, unlinkfile bool) (bool, *EjdbError) {
    c_colname := C.CString(colname)
    defer C.free(unsafe.Pointer(c_colname))
    res := C.ejdbrmcoll((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_colname, C._Bool(unlinkfile))
    if res {
        return bool(res), nil
    }
    return bool(res), ejdb.check_error()
}

// Synchronize entire EJDB database and all of its collections with storage.
func (ejdb *Ejdb) Sync() (bool, *EjdbError) {
    ret := C.ejdbsyncdb((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)))
    if ret {
        return bool(ret), nil
    }
    return bool(ret), ejdb.check_error()
}

// Gets description of EJDB database and its collections as a BSON object.
func (ejdb *Ejdb) Meta() ([]byte, *EjdbError) {
    bson := C.ejdbmeta((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)))
    err := ejdb.check_error()
    if err != nil {
        return make([]byte, 0), err
    }
    defer C.bson_del(bson)
    return bson_to_byte_slice(bson), nil
}

// Imports previously exported collections data into ejdb.
// Global database write lock will be applied during import operation.
//
// NOTE: Only data exported as BSONs can be imported with `Ejdb.Import()`
//
// `path` is the directory path in which data resides.
// `cnames` is a list of collection names to import. `nil` implies that all collections found in `path` will be imported.
// `flags` can be one of:
//             `JBIMPORTUPDATE`  Update existing collection entries with imported ones.
//                               Existing collections will not be recreated.
//                               For existing collections options will not be imported.
//
//             `JBIMPORTREPLACE` Recreate existing collections and replace
//                               all their data with imported entries.
//                               Collections options will be imported.
//
//             `0`              Implies `JBIMPORTUPDATE`
// Import() returns the log for the operation as a string
func (ejdb *Ejdb) Import(path string, cnames *[]string, flags int) (log string, err *EjdbError) {
    c_path := C.CString(path)
    defer C.free(unsafe.Pointer(c_path))

    c_log := C.tcxstrnew()
    defer C.tcxstrdel(c_log)

    if cnames != nil {
        tclist := C.tclistnew2(C.int(len(*cnames)))
        defer C.tclistdel(tclist)
        for i := 0; i < len(*cnames); i++ {
            cname := C.CString((*cnames)[i])
            defer C.free(unsafe.Pointer(cname))
            C.tclistpush2(tclist, cname)
        }

        C.ejdbimport((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_path, tclist, C.int(flags), c_log)
    } else {
        C.ejdbimport((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_path, nil, C.int(flags), c_log)
    }

    c_chars := C.tcxstrptr(c_log)
    ret := C.GoString((*C.char)(c_chars))

    return ret, ejdb.check_error()
}

// Exports database collections data to the specified directory.
// Database read lock will be taken on each collection.
//
// NOTE: Only data exported as BSONs can be imported with `Ejdb.Import()`
//
// `path` is the directory path in which data will be exported.
// `cnames` is a list of collection names to export. `nil` implies that all existing collections will be exported.
// `flags` can be set to `JBJSONEXPORT` in order to export data as JSON files instead exporting into BSONs.
// Export() returns the log for the operation as a string
func (ejdb *Ejdb) Export(path string, cnames *[]string, flags int) (log string, err *EjdbError) {
    c_path := C.CString(path)
    defer C.free(unsafe.Pointer(c_path))

    c_log := C.tcxstrnew()
    defer C.tcxstrdel(c_log)

    if cnames != nil {
        tclist := C.tclistnew2(C.int(len(*cnames)))
        defer C.tclistdel(tclist)
        for i := 0; i < len(*cnames); i++ {
            cname := C.CString((*cnames)[i])
            defer C.free(unsafe.Pointer(cname))
            C.tclistpush2(tclist, cname)
        }

        C.ejdbexport((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_path, tclist, C.int(flags), c_log)
    } else {
        C.ejdbexport((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_path, nil, C.int(flags), c_log)
    }

    c_chars := C.tcxstrptr(c_log)
    ret := C.GoString((*C.char)(c_chars))

    return ret, ejdb.check_error()
}

// Execute ejdb database command.
//
// Supported commands:
//
//
//  1) Exports database collections data. See ejdbexport() method.
//
//    "export" : {
//          "path" : string,                    //Exports database collections data
//          "cnames" : [string array]|null,     //List of collection names to export
//          "mode" : int|null                   //Values: null|`JBJSONEXPORT` See ejdbexport() method
//    }
//
//    Command response:
//       {
//          "log" : string,        //Diagnostic log about executing this command
//          "error" : string|null, //ejdb error message
//          "errorCode" : int|0,   //ejdb error code
//       }
//
//  2) Imports previously exported collections data into ejdb.
//
//    "import" : {
//          "path" : string                     //The directory path in which data resides
//          "cnames" : [string array]|null,     //List of collection names to import
//          "mode" : int|null                //Values: null|`JBIMPORTUPDATE`|`JBIMPORTREPLACE` See ejdbimport() method
//     }
//
//     Command response:
//       {
//          "log" : string,        //Diagnostic log about executing this command
//          "error" : string|null, //ejdb error message
//          "errorCode" : int|0,   //ejdb error code
//       }
func (ejdb *Ejdb) Command(bson []byte) (*[]byte, *EjdbError) {
    c_bson := bson_from_byte_slice(bson)
    defer C.bson_destroy(c_bson)
    return ejdb.command(c_bson)
}

func (ejdb *Ejdb) JsonCommand(json string) (*[]byte, *EjdbError) {
    c_bson := bson_from_json(json)
    defer C.bson_destroy(c_bson)
    return ejdb.command(c_bson)
}

func (ejdb *Ejdb) command(c_bson *C.bson) (*[]byte, *EjdbError) {
    out_c_bson := C.ejdbcommand((*C.struct_EJDB)(unsafe.Pointer(ejdb.ptr)), c_bson)
    if out_c_bson == nil {
        return nil, ejdb.check_error()
    }
    defer C.bson_del(out_c_bson)
    out_bson := bson_to_byte_slice(out_c_bson)
    return &out_bson, nil
}
