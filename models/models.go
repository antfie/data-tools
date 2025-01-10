package models

import "gorm.io/gorm"

type PathHash struct {
	ID      uint   `gorm:"primarykey"`
	Hash    string `gorm:"unique"`
	Ignored bool
	Size    *uint
}

type Path struct {
	ID             uint `gorm:"primarykey"`
	ParentPathID   *uint
	ParentPath     *Path
	Level          uint
	Name           string
	ChildPathCount *uint
	PathHashID     *uint
	PathHash       *PathHash
	Ignored        bool
	Size           *uint
	DeletedAt      gorm.DeletedAt
}

type FileType struct {
	ID   uint   `gorm:"primarykey"`
	Type string `gorm:"unique"`
}

type FileHash struct {
	ID         uint   `gorm:"primarykey"`
	Hash       string `gorm:"unique"`
	Ignored    bool
	Size       *uint
	FileTypeID *uint
	FileType   *FileType
	Zapped     bool
}

type File struct {
	ID         uint `gorm:"primarykey"`
	PathID     uint
	Path       Path
	Level      uint
	FileHashID *uint
	FileHash   *FileHash
	Name       string
	Size       *uint
	FileTypeID *uint
	FileType   *FileType
	Ignored    bool
	Zapped     bool
	DeletedAt  gorm.DeletedAt
}

type Note struct {
	gorm.Model
	Note string
}
type PathHashNote struct {
	gorm.Model
	PathHashID uint
	PathHash   PathHash
	NoteId     uint
	Note       Note
}

type PathNote struct {
	gorm.Model
	PathID uint
	Path   Path
	NoteId uint
	Note   Note
}

type FileTypeNote struct {
	gorm.Model
	FileTypeID uint
	FileType   FileType
	NoteId     uint
	Note       Note
}

type FileHashNote struct {
	gorm.Model
	FileHashID uint
	FileHash   FileHash
	NoteId     uint
	Note       Note
}

type FileNote struct {
	gorm.Model
	FileID uint
	File   File
	NoteId uint
	Note   Note
}
