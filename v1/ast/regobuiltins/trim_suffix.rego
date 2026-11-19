package system.regobuiltins.trim_suffix

impl(s, suffix) := substring(s, 0, count(s) - count(suffix)) if {
	endswith(s, suffix)
} else := s
