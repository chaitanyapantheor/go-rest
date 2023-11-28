package buildgrp

import "github.com/chaitanyamaili/go_rest/models/build"

// swagger:response BuildRes
type _ struct {
	// in:body
	Body struct {
		// Success
		//
		Success bool `json:"success"`
		// Timestamp
		//
		// example: 1639237536
		Timestamp int64 `json:"timestamp"`
		// Data
		// in: body
		Data []build.Build `json:"data"`
	}
}

// swagger:parameters BuildQueryById
type _ struct {
	// Build ID
	//
	// in: path
	// required: true
	// enum: 1
	// type: integer
	ID string `json:"id"`
}

// swagger:parameters BuildCreate
type _ struct {
	// Build input Json Object
	//
	// in: body
	// required: true
	Body build.NewBuild
}
