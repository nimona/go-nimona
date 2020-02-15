package immutable

func (v List) IsList() bool   { return true }
func (v List) IsMap() bool    { return false }
func (v List) IsBool() bool   { return false }
func (v List) IsString() bool { return false }
func (v List) IsInt() bool    { return false }
func (v List) IsFloat() bool  { return false }
func (v List) IsBytes() bool  { return false }

func (v Map) IsList() bool   { return false }
func (v Map) IsMap() bool    { return true }
func (v Map) IsBool() bool   { return false }
func (v Map) IsString() bool { return false }
func (v Map) IsInt() bool    { return false }
func (v Map) IsFloat() bool  { return false }
func (v Map) IsBytes() bool  { return false }

func (v Bool) IsList() bool   { return false }
func (v Bool) IsMap() bool    { return false }
func (v Bool) IsBool() bool   { return true }
func (v Bool) IsString() bool { return false }
func (v Bool) IsInt() bool    { return false }
func (v Bool) IsFloat() bool  { return false }
func (v Bool) IsBytes() bool  { return false }

func (v String) IsList() bool   { return false }
func (v String) IsMap() bool    { return false }
func (v String) IsBool() bool   { return false }
func (v String) IsString() bool { return true }
func (v String) IsInt() bool    { return false }
func (v String) IsFloat() bool  { return false }
func (v String) IsBytes() bool  { return false }

func (v Int) IsList() bool   { return false }
func (v Int) IsMap() bool    { return false }
func (v Int) IsBool() bool   { return false }
func (v Int) IsString() bool { return false }
func (v Int) IsInt() bool    { return true }
func (v Int) IsFloat() bool  { return false }
func (v Int) IsBytes() bool  { return false }

func (v Float) IsList() bool   { return false }
func (v Float) IsMap() bool    { return false }
func (v Float) IsBool() bool   { return false }
func (v Float) IsString() bool { return false }
func (v Float) IsInt() bool    { return false }
func (v Float) IsFloat() bool  { return true }
func (v Float) IsBytes() bool  { return false }

func (v Bytes) IsList() bool   { return false }
func (v Bytes) IsMap() bool    { return false }
func (v Bytes) IsBool() bool   { return false }
func (v Bytes) IsString() bool { return false }
func (v Bytes) IsInt() bool    { return false }
func (v Bytes) IsFloat() bool  { return false }
func (v Bytes) IsBytes() bool  { return true }
