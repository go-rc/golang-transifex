golang-transifex
================

Simple library and command line tools for uploading and downloading translation files to/from transifex server

Upload
------

Upload all the 'source language' (default english) files defined in the configuration file.
Optionally also upload translations.

Transifex has the concept of "resources".  Each resource represents a set of translation files for all languages.  There is a source language which contains all strings that need to be translated.  This file is not editable.  There are also any number of "translations" for a "resource".  Each translation are all strings in a particular language that have been translated.

Running the upload command will:

1. Create all resources in the configuration file that are not currently in transifex
2. Upload the contents of the 'source language' translations file.
3. If a resource was created all translations will be uploaded


Download
--------

All translations will be downloaded and written to the appropriate files.  This will overwrite the previous files on disk without warning but if there is a problem git can be used to roll back the changes.

Configuration
-------------

A configuration file is required that declares all the "resources".  

	{
		"KEYVALUEJSON":[{
			"dir" : "web-ui/src/main/resources/catalog/locales/",
			"filename" : "core",
			"name" : "Angular UI Common Strings",
			"slug" : "core",
			"priority" : "0",
			"categories" : ["Angular_UI"]
		}]
	}

* the key of the configuration object indicates the format of the translations files.  
* `dir` is the directory containing the translation files
* `filename` is the used to find the translation files.  This is required if there are several sets of translation files in the same directory.  For example: 
 * en-admin.json, fr-admin.json
 * en-core.json, en-core.json
* `name` the display name in transifex of the resource
* `slug` an id for the resource in transifex.  This is used to address the resource via the transifex API.
* `priority` importance of the resource.  
* `categories` categories of the resource to use for organizing translation files