package permissionutil

var (
	IsRoot       bool
	HasCapNetRaw bool
)

const (
	os_read        = 04
	os_write       = 02
	os_ex          = 01
	os_user_shift  = 6
	os_group_shift = 3
	os_other_shift = 0

	// User Read Write Execute Permission
	UserRead             = os_read << os_user_shift
	UserWrite            = os_write << os_user_shift
	UserExecute          = os_ex << os_user_shift
	UserReadWrite        = UserRead | UserWrite
	UserReadWriteExecute = UserReadWrite | UserExecute

	// Group Read Write Execute Permission
	GroupRead             = os_read << os_group_shift
	GroupWrite            = os_write << os_group_shift
	GroupExecute          = os_ex << os_group_shift
	GroupReadWrite        = GroupRead | GroupWrite
	GroupReadWriteExecute = GroupReadWrite | GroupExecute

	// Other Read Write Execute Permission
	OtherRead             = os_read << os_other_shift
	OtherWrite            = os_write << os_other_shift
	OtherExecute          = os_ex << os_other_shift
	OtherReadWrite        = OtherRead | OtherWrite
	OtherReadWriteExecute = OtherReadWrite | OtherExecute

	// All Read Write Execute Permission
	AllRead             = UserRead | GroupRead | OtherRead
	AllWrite            = UserWrite | GroupWrite | OtherWrite
	AllExecute          = UserExecute | GroupExecute | OtherExecute
	AllReadWrite        = AllRead | AllWrite
	AllReadWriteExecute = AllReadWrite | AllExecute
)

func init() {
	IsRoot, _ = checkCurrentUserRoot()
	HasCapNetRaw, _ = checkCurrentUserCapNetRaw()
}
