package controllers

import jsoniter "github.com/json-iterator/go"

// json is the shared JSON codec for all controllers, compatible with the standard library.
var json = jsoniter.ConfigCompatibleWithStandardLibrary
