### Internationalization & Localization

## Example locale

`locale/en.toml` =

```
[datetime]
	[distance_in_words]
		half_a_minute = "half a minute"
		[about_x_hours]
			one = "about 1 hour"
			other = "about %s hours"
		[about_x_months]
			one = "about 1 month"
			other = "about %s months"
		[about_x_years]
			one = "about 1 year"
			other = "about %s years"
		[almost_x_years]
			one = "almost 1 year"
			other = "almost %s years"
		[less_than_x_minutes]
			one = "less than a minute"
			other = "less than %s minutes"
		[less_than_x_seconds]
			one = "less than 1 second"
			other = "less than %s seconds"
		[over_x_years]
			one = "over 2 year"
			other = "over %s years"
		[x_days]
			one = "1 day"
			other = "%s days"
		[x_minutes]
			one = "1 minute"
			other = "%s minutes"
		[x_months]
			one = "1 month"
			other = "%s months"
		[x_years]
			one = "1 year"
			other = "%s years"  
		[x_seconds]
			one = "1 second"
			other = "%s seconds"
[number]
	[regular]
		delimiter = ','
		precision = 3
		seperator = '.'
		strip_insigificant_zeros = false
	[decimal]
		format = '%n %u'
		precision = 3
		strip_insigificant_zeros = true
		units = ['Hundred', 'Thousand', 'Million', 'Billion', 'Trillion', 'Quadrillion']
	[currency]
		format = '%u%n'
		delimiter = ','
		unit = '$'
		precision = 2
		seperator = '.'
		strip_insigificant_zeros = false
	[percentage]
		format = '%u%n'
		delimiter = ''
		precision = 2
		unit = '%'
		strip_insigificant_zeros = false
	[storage]
		format = '%n %u'
		precision = 3
		strip_insigificant_zeros = true
		units = ['B', 'KB', 'MB', 'GB', 'TB']
[date]
	days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']	
	months = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December']
	[abbreviations]
		days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
		months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
	[formats]
		default = ''
		short = ''
		long = ''
[button]
	download = 'Download'
	upload = 'Upload'
	create = 'Create'
	submit = 'Submit'
	update = 'Update'
	delete = 'Delete'
	remove = 'Remove'
	cancel = 'Cancel'
	select = 'Select'
	approve = 'Approve'
	okay = 'Okay'
	next = 'Next'
	prev = 'Prev'
	back = 'Back'
	undo = 'Undo'
	redo = 'Redo'
	apply = 'Apply'
	close = 'Close'
	list = 'List'
[time]
	am = 'am'
	pm = 'pm'
	[formats]
		default = ''
		short = ''
		long = ''
[error]
	accepted = "%s must be accepted"
	blank = "%s can't be blank"
	present = "%s must be blank"
	confirmation = "%s doesn't match %s"
	empty = "%s can't be empty"
	equal_to = "must be equal to %s"
	even = "%s must be even"
	exclusion = "%s is reserved"
	greater_than = "%s must be greater than %s"
	greater_than_or_equal_to = "%s must be greater than or equal to %s"
	inclusion = "%s is not included in the list"
	invalid = "%s is invalid"
	less_than = "%s must be less than %s"
	less_than_or_equal_to = "%s must be less than or equal to %s"
	model_invalid = "Validation failed: %s"
	not_a_number = "%s is not a number"
	not_an_integer = "%s must be an integer"
	odd = "%s must be odd"
	required = "%s must exist"
	taken = "%s has already been taken"
	too_long = "%s is too long (maximum is %s characters)"
	too_short = "%s is too short (minimum is %s characters)"
	wrong_length = "%s is the wrong length (should be %s characters)"
	other_than = "%s must be other than %s"
	template_body = "There were problems with the following fields: %s"
	template_header = "%s errors prohibited changes being made"
```

### Implementation spec (TODO)

```
type LocaleFinder interface {
  DetermineLocale(r *http.Request) string
}

type LocalizedRenderData string

func (l LocalizedRenderData) Localize(s string) {
  // look up stuff in thing.
}

func MkLocale(l LocaleFinder, r *http.Request) LocalizedRenderData {
   return LocalizedRenderData{
     Locale: l.DetermineLocale(r),
     Data: nil,
   }
}
```

```
LoadLocales loads the locales into the app state object at the start of the app lifecycle to load the locales into memory doing a merge, 
it iterates over a known folder with toml locale files in it (en-US.toml etc), and the filename becomes the key of the map, and the map 
is a map of string string and you look up the locale and you get a table of strings back.

panic if can't find key.

locale/en.toml
locale/en-US.toml
locale/en-CA.toml

If can't find translation in en-US.toml for example, fall back to parent which is en.toml
```

```
Need to load the configuration for internationalization.
Probably have a directory of things that say:
	en.toml
	en-US.toml
	
i18n.Load()

Needs to be able to take an http.Request and determine the locale. Should use header detection first, fall back to en-us.

* Check what the semantic meaning of locale identifiers is
* Locale Determiner Function
* Middleware to set the locale in the context
```
