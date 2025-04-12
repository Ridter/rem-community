package core

import (
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"

	"github.com/chainreactors/rem/x/utils"
	"gopkg.in/yaml.v3"
)

var AvailableWrappers []string

func ParseWrapperOptions(s string, key string) (WrapperOptions, error) {
	decodedData, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	decryptedData, err := utils.AesDecrypt(decodedData, utils.PKCS7Padding([]byte(key), 16))
	if err != nil {
		return nil, err
	}

	var options WrapperOptions
	err = yaml.Unmarshal(decryptedData, &options)
	if err != nil {
		return nil, err
	}

	return options, nil
}

type WrapperOptions []*WrapperOption

func (opts WrapperOptions) String(key string) string {
	marshal, err := yaml.Marshal(opts)
	if err != nil {
		utils.Log.Error(err)
		return ""
	}
	data, err := utils.AesEncrypt(marshal, utils.PKCS7Padding([]byte(key), 16))
	if err != nil {
		utils.Log.Error(err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(data)
}

// WrapperOption 定义wrapper的基本配置结构
type WrapperOption struct {
	Name    string            `yaml:"name"`
	Options map[string]string `yaml:"options"`
}

// Wrapper 定义了数据包装器的基本接口
type Wrapper interface {
	Name() string
	io.ReadWriteCloser
}

// WrapperCreatorFn 是创建wrapper的工厂函数类型
type WrapperCreatorFn func(r io.Reader, w io.Writer, opt map[string]string) (Wrapper, error)

var wrapperCreators = make(map[string]WrapperCreatorFn)

// WrapperRegister 注册一个wrapper类型
func WrapperRegister(name string, fn WrapperCreatorFn) {
	if _, exist := wrapperCreators[name]; exist {
		panic(fmt.Sprintf("wrapper [%s] is already registered", name))
	}
	wrapperCreators[name] = fn
}

// WrapperCreate 创建一个指定类型的wrapper
func WrapperCreate(name string, r io.Reader, w io.Writer, opt map[string]string) (Wrapper, error) {
	if fn, ok := wrapperCreators[name]; ok {
		return fn(r, w, opt)
	}
	return nil, fmt.Errorf("wrapper [%s] is not registered", name)
}

// GenerateRandomWrapperOption 生成单个随机的WrapperOption
func GenerateRandomWrapperOption() *WrapperOption {
	name := AvailableWrappers[rand.Intn(len(AvailableWrappers))]

	opt := &WrapperOption{
		Name:    name,
		Options: make(map[string]string),
	}

	switch name {
	case AESWrapper:
		opt.Options["key"] = utils.RandomString(32)
		opt.Options["iv"] = utils.RandomString(16)
	case XORWrapper:
		opt.Options["key"] = utils.RandomString(32)
		opt.Options["iv"] = utils.RandomString(16)
	}

	return opt
}

// GenerateRandomWrapperOptions 生成随机数量、随机组合的Wrapper配置
func GenerateRandomWrapperOptions(minCount, maxCount int) WrapperOptions {
	if minCount < 1 {
		minCount = 1
	}
	if maxCount < minCount {
		maxCount = minCount
	}

	count := minCount
	if maxCount > minCount {
		count += rand.Intn(maxCount - minCount + 1)
	}

	var opts WrapperOptions
	for i := 0; i < count; i++ {
		opt := GenerateRandomWrapperOption()
		opts = append(opts, opt)
	}

	rand.Shuffle(len(opts), func(i, j int) {
		opts[i], opts[j] = opts[j], opts[i]
	})

	return opts
}
