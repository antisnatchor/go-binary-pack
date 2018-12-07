# Go BinaryPack

[![Build Status](https://travis-ci.org/roman-kachanovsky/go-binary-pack.svg?branch=master)](https://travis-ci.org/roman-kachanovsky/go-binary-pack)
[![GoDoc](https://godoc.org/github.com/roman-kachanovsky/go-binary-pack/binary-pack?status.svg)](http://godoc.org/github.com/roman-kachanovsky/go-binary-pack/binary-pack)

BinaryPack is a simple Golang library which implements some functionality of Python's [struct](https://docs.python.org/2/library/struct.html) package.

This is a fork of Roman Kachanovsky [go-binary-pack](https://github.com/roman-kachanovsky/go-binary-pack)

**Note**

binary_pack performs conversions between some Go values represented as byte slices.
	This can be used in handling binary data stored in files or from network connections,
	among other sources. It uses format slices of strings as compact descriptions of the layout
	of the Go structs.

	Format characters (some characters like H have been reserved for future implementation of unsigned numbers):
		? - bool, packed size 1 byte
		B - int, packed size 1 byte
		h, H - int, packed size 2 bytes (in future it will support pack/unpack of int8, uint8 values)
		i, I, l, L - int, packed size 4 bytes (in future it will support pack/unpack of int16, uint16, int32, uint32 values)
		q, Q - int, packed size 8 bytes (in future it will support pack/unpack of int64, uint64 values)
		f - float32, packed size 4 bytes
		d - float64, packed size 8 bytes
		Ns - string, packed size N bytes, N is a number of runes to pack/unpack
		
This code was hacked during @marver and @antisnatchor WebUSB research in March 2018.
More info here: [A Journey into Novel Web-Technology and U2F Exploitation](https://www.youtube.com/watch?v=pUa6nWWTO4o)

Main changes from the original project:
 - defaults to BigEndian, with auto-switch to LittleEndian dpending on < and > usage
 - added unsigned char (B) as a 1 byte integer
 - added more advanced usage examples

**Install**

`go get github.com/antisnatchor/go-binary-pack/binary-pack`

**Basic Usage**

```go
// Prepare format (slice of strings)
format := []string{"I", "?", "d", "6s"}

// Prepare values to pack
values := []interface{}{4, true, 3.14, "Golang"}

// Create BinaryPack object
bp := new(BinaryPack)

// Pack values to []byte
data, err := bp.Pack(format, values)

// Unpack binary data to []interface{}
unpacked_values, err := bp.UnPack(format, data)

// You can calculate size of expected binary data by format
size, err := bp.CalcSize(format)

```


**More advanced example taken from the WebUSB research**

Define the data structure:

```go

// define the data structure
type OPRETDevList struct {
	Version uint16
	Command uint16
	Status  uint32

	ExpDev              uint32
	Path                string
	BusID               string
	BusNum              uint32
	DevNum              uint32
	Speed               uint32
	IsVendor            uint16
	IsProduct           uint16
	BcdDevice           uint16
	BDeviceClass        uint8
	BDeviceSubClass     uint8
	BDeviceProtocol     uint8
	BConfigurationValue uint8
	BNumConfigurations  uint8
	BNumInterfaces      uint8

	BInterfaceClass    uint8
	BInterfaceSubClass uint8
	BInterfaceProtocol uint8
	Align              uint8
}
```


Define the format of the data in the structure:

```go
// header
var FormatUSBIPHeader = []string{
	"H", //('Version', 'H', 273),
	"H", //('Command uint16
	"I", //('Status', 'I')
}

// trailer
var FormatUSBInterface = []string{
	"B", //('BInterfaceClass uint8
	"B", //('BInterfaceSubClass uint8
	"B", //('BInterfaceProtocol uint8
	"B", //('Align', 'B', 0)
}

// main embedding header and trailer
var FormatOPREpDevList = append( // ugly trick we know ;)
	append(
		FormatUSBIPHeader,
		[]string{
			"I",    //('nExportedDevice uint32
			"256s", //('UsbPath', '256s'),
			"32s",  //('BusID', '32s'),
			"I",    //('BusNum uint32
			"I",    //('DevNum uint32
			"I",    //('Speed uint32
			"H",    //('IdVendor uint16
			"H",    //('IdProduct uint16
			"H",    //('BcdDevice uint16
			"B",    //('BDeviceClass uint8
			"B",    //('BDeviceSubClass uint8
			"B",    //('BDeviceProtocol uint8
			"B",    //('BConfigurationValue uint8
			"B",    //('BNumConfigurations uint8
			"B",    //('BNumInterfaces uint8
		}...),
	FormatUSBInterface...,
)

```
Define the data using the structure created above:

```go
devList := OPRETDevList{
    Version: 273,
    Command: OpRepDevList,
    Status:  0,

    ExpDev:              1,
    Path:                "/sys/devices/pci0000:00/0000:00:01.2/usb1/1-1",
    BusID:               "1-1",
    BusNum:              1,
    DevNum:              0x1b,
    Speed:               2,
    IsVendor:            h.IDVendor,
    IsProduct:           h.IDProduct,
    BcdDevice:           0,
    BDeviceClass:        0,
    BDeviceSubClass:     0,
    BDeviceProtocol:     0,
    BConfigurationValue: 0,
    BNumConfigurations:  1,
    BNumInterfaces:      1,
    BInterfaceClass:    11,
    BInterfaceSubClass: 0,
    BInterfaceProtocol: 0,
    Align:              0,
}
```

Finally wrap the Pack functionality and call it:

```go
func (instance OPRETDevList) Pack() []byte {
	values := []interface{}{
		instance.Version,
		instance.Command,
		instance.Status,

		instance.ExpDev,
		instance.Path,
		instance.BusID,
		instance.BusNum,
		instance.DevNum,
		instance.Speed,
		instance.IsVendor,
		instance.IsProduct,
		instance.BcdDevice,
		instance.BDeviceClass,
		instance.BDeviceSubClass,
		instance.BDeviceProtocol,
		instance.BConfigurationValue,
		instance.BNumConfigurations,
		instance.BNumInterfaces,

		//Interfaces
		instance.BInterfaceClass,
		instance.BInterfaceSubClass,
		instance.BInterfaceProtocol,
		instance.Align,
	}
	bp := new(binarypack.BinaryPack)
	// Pack values to []byte
	data, err := bp.Pack(FormatOPREpDevList, values)

	handleErr(err)

	return data
}

// call Pack and enjoy properly structured typed binary data out
data := devList.Pack()


```
