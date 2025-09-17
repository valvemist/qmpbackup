## qmpbackup

qmpbackup is a Go-based command-line tool for orchestrating QEMU block device live backups using the QEMU Machine Protocol (QMP).

**Note: This project is unrelated to [abbbi/qmpbackup](https://github.com/abbbi/qmpbackup).**  

## Features

- Full and incremental backup support

## Usage

To run a backup:

    ./backup -socket /tmp/qemu-qmp-socket -backupFile /path/to/image.qcow2 -deviceToBackup drive0

Required Parameters:
```
 -socket           Path to the QMP UNIX socket (e.g. /tmp/qemu-qmp-socket)  
 -backupFile       Path to the QCOW2 image file to write the backup  
 -deviceToBackup   QEMU block device name (e.g. drive0 from "-device scsi-hd,drive=drive0")  
```

Optional Flags:
```
  -inc  int        Incremental level (-1 means full backup) (default -1)  
  -cleanBitmap     If set, removes the bitmap and exits without performing a backup  
  -v               If set, verbose output showing used JSONs  
```

When performing incremental backups, the tool automatically generates filenames by appending level-specific suffixes to the path provided via -backupFile.

```
 /tmp/drive0.full.qcow2
 /tmp/drive0.inc0.qcow2
 /tmp/drive0.inc1.qcow2
 /tmp/drive0.inc2.qcow2
```

If the specified full backup qcow2 file does not exist, it will be created automatically. However, attempting to perform a full backup to an existing non-empty file will result in failure.
Additionally, incremental backups will silently fail if the target incremental backup file already exists.

Workflow Overview
-----------------

1. Connect to QEMU via QMP
2. Add block device (create image if missing)
3. Add bitmap (if incremental)
4. Perform backup
5. Listen for completion events

Documentation
-------------

- All exported functions are documented inline
- See qmpbackup/doc.go and cmd/backup/doc.go for package-level descriptions

Build Instructions
------------------

This project is written in Go and uses Go modules for dependency management.

To build the CLI tool:

  1. Make sure you have Go 1.20 or newer installed.
     You can check your version with:
       go version

  2. Clone the repository:
       git clone https://github.com/valvemist/qmpbackup.git
       cd qmpbackup

  3. Build the binary:
       go build -o qmpbackup ./cmd/backup

     This will create an executable named 'qmpbackup' in the current directory.

To run the tool directly without building:

       go run ./cmd/backup \
         -socket /tmp/qemu-qmp-socket \
         -backupFile /path/to/image.qcow2 \
         -deviceToBackup drive0

Note: The tool requires a running QEMU instance with QMP enabled and a valid socket path.

Testing
-------

TESTING NOT YET IMPLEMENTED!  
To run tests:
  go test ./qmpbackup/...

> ⚠️ **Beta Notice**: This project is currently in beta. Features and APIs may change, and stability is not guaranteed.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.

## Third-Party Libraries

This project uses the following dependencies:

- [go-qemu/qmp](https://github.com/digitalocean/go-qemu) — Apache 2.0
- [gjson](https://github.com/tidwall/gjson) — MIT
- [pretty](https://github.com/tidwall/pretty) — MIT
- [sjson](https://github.com/tidwall/sjson) — MIT

See [THIRD_PARTY_LICENSES](./THIRD_PARTY_LICENSES) for full license texts.
