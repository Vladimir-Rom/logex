package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

type Registry struct {
	k  *koanf.Koanf
	fs *pflag.FlagSet
}

func NewRegistry(k *koanf.Koanf, fs *pflag.FlagSet) *Registry {
	return &Registry{
		k:  k,
		fs: fs,
	}
}

func (r *Registry) IntP(name, shorthand string, defaultValue int, help string) func() int {
	return defineParamP(
		name,
		shorthand,
		defaultValue,
		help,
		r.fs.IntP,
		r.k.Int,
	)
}

func (r *Registry) Int(name string, defaultValue int, help string) func() int {
	return defineParam(
		name,
		defaultValue,
		help,
		r.fs.Int,
		r.k.Int,
	)
}

func (r *Registry) BoolP(name, shorthand string, defaultValue bool, help string) func() bool {
	return defineParamP(
		name,
		shorthand,
		defaultValue,
		help,
		r.fs.BoolP,
		r.k.Bool,
	)
}

func (r *Registry) Bool(name string, defaultValue bool, help string) func() bool {
	return defineParam(
		name,
		defaultValue,
		help,
		r.fs.Bool,
		r.k.Bool,
	)
}

func (r *Registry) StringP(name, shorthand, defaultValue, help string) func() string {
	return defineParamP(
		name,
		shorthand,
		defaultValue,
		help,
		r.fs.StringP,
		r.k.String,
	)
}

func (r *Registry) String(name, defaultValue, help string) func() string {
	return defineParam(
		name,
		defaultValue,
		help,
		r.fs.String,
		r.k.String,
	)
}

func (r *Registry) StringsP(name, shorthand string, defaultValue []string, help string) func() []string {
	return defineParamP(
		name,
		shorthand,
		defaultValue,
		help,
		r.fs.StringSliceP,
		r.k.Strings,
	)
}

func (r *Registry) Strings(name string, defaultValue []string, help string) func() []string {
	return defineParam(
		name,
		defaultValue,
		help,
		r.fs.StringSlice,
		r.k.Strings,
	)
}

func defineParam[T any](
	name string,
	defaultValue T,
	help string,
	implFlag func(name string, defaultValue T, help string) *T,
	implConf func(name string) T) func() T {
	implFlag(
		name,
		defaultValue,
		help)
	return func() T {
		return implConf(name)
	}
}

func defineParamP[T any](
	name, shorthand string,
	defaultValue T,
	help string,
	implFlag func(name, shorthand string, defaultValue T, help string) *T,
	implConf func(name string) T) func() T {
	implFlag(
		name,
		shorthand,
		defaultValue,
		help)
	return func() T {
		return implConf(name)
	}
}
