package somatree

import "github.com/satori/go.uuid"

type SomaTreeProperty interface {
	GetType() string
	GetID() string
	GetSource() string
	hasInheritance() bool
	isChildrenOnly() bool
}

//
// Custom
type SomaTreePropertyCustom struct {
	Id            uuid.UUID
	Inherited     bool
	InheritedFrom uuid.UUID
	Inheritance   bool
	ChildrenOnly  bool
	View          string
	Key           string
	value         string
}

func (p *SomaTreePropertyCustom) GetType() string {
	return "custom"
}

func (p *SomaTreePropertyCustom) GetID() string {
	return p.Id.String()
}

func (p *SomaTreePropertyCustom) GetSource() string {
	return p.InheritedFrom.String()
}

func (p *SomaTreePropertyCustom) hasInheritance() bool {
	return p.Inheritance
}

func (p *SomaTreePropertyCustom) isChildrenOnly() bool {
	return p.ChildrenOnly
}

//
// Service
type SomaTreePropertyService struct {
	Id            uuid.UUID
	Inherited     bool
	InheritedFrom uuid.UUID
	Inheritance   bool
	ChildrenOnly  bool
	View          string
	Service       string
	Attributes    []SomaTreePropertyServiceAttribute
}

type SomaTreePropertyServiceAttribute struct {
	Attribute string
	Value     string
}

func (p *SomaTreePropertyService) GetType() string {
	return "service"
}

func (p *SomaTreePropertyService) GetID() string {
	return p.Id.String()
}

func (p *SomaTreePropertyService) GetSource() string {
	return p.InheritedFrom.String()
}

func (p *SomaTreePropertyService) hasInheritance() bool {
	return p.Inheritance
}

func (p *SomaTreePropertyService) isChildrenOnly() bool {
	return p.ChildrenOnly
}

//
// System
type SomaTreePropertySystem struct {
	Id            uuid.UUID
	Inherited     bool
	InheritedFrom uuid.UUID
	Inheritance   bool
	ChildrenOnly  bool
	View          string
	Key           string
	Value         string
}

func (p *SomaTreePropertySystem) GetType() string {
	return "system"
}

func (p *SomaTreePropertySystem) GetID() string {
	return p.Id.String()
}

func (p *SomaTreePropertySystem) GetSource() string {
	return p.InheritedFrom.String()
}

func (p *SomaTreePropertySystem) hasInheritance() bool {
	return p.Inheritance
}

func (p *SomaTreePropertySystem) isChildrenOnly() bool {
	return p.ChildrenOnly
}

//
// Oncall
type SomaTreePropertyOncall struct {
	Id            uuid.UUID
	Inherited     bool
	InheritedFrom uuid.UUID
	Inheritance   bool
	ChildrenOnly  bool
	View          string
	Oncall        string
}

func (p *SomaTreePropertyOncall) GetType() string {
	return "return oncall"
}

func (p *SomaTreePropertyOncall) GetID() string {
	return p.Id.String()
}

func (p *SomaTreePropertyOncall) GetSource() string {
	return p.InheritedFrom.String()
}

func (p *SomaTreePropertyOncall) hasInheritance() bool {
	return p.Inheritance
}

func (p *SomaTreePropertyOncall) isChildrenOnly() bool {
	return p.ChildrenOnly
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
