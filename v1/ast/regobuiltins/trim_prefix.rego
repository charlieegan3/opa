package system.regobuiltins.trim_prefix

impl(s, prefix) := substring(s, count(prefix), -1) if {
	substring(s, 0, count(prefix)) == prefix
} else := s
