package cryptutil

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/GehirnInc/crypt/md5_crypt"
	"github.com/GehirnInc/crypt/sha256_crypt"
	"github.com/GehirnInc/crypt/sha512_crypt"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

func Sha512Crypt(salt string, content string) (string, error) {

	//password := "my_password"
	//salt := "mysalt123"

	// crypt 格式
	hash := fmt.Sprintf("$6$%s", salt)
	c := sha512_crypt.New()
	result, err := c.Generate([]byte(content), []byte(hash))
	if err != nil {
		// 程序遇到严重错误，无法继续运行，立即停止执行并输出错误信息和调用栈。程序不会终止
		panic(err)
	}

	fmt.Println("Hash:", result)

	// 验证密码
	/*err = crypt.CompareHashAndPassword(result, []byte(password))
	if err != nil {
		fmt.Println("Password not match")
	} else {
		fmt.Println("Password match")
	}*/
	return string(result), nil
}

func Sha256Crypt(salt string, content string) (string, error) {
	//cryptSHA256 := crypt.SHA256.New()
	cryptSHA256 := sha256_crypt.New()
	ret, _ := cryptSHA256.Generate([]byte(content), []byte("$5$"+salt))
	fmt.Println(ret)

	err := cryptSHA256.Verify(ret, []byte(content))
	fmt.Println(err)
	if err != nil {
		return "", err
	}

	// Output:
	// $5$salt$kpa26zwgX83BPSR8d7w93OIXbFt/d3UOTZaAu5vsTM6
	// <nil>
	return ret, err
}

func Md5Crypt(salt string, content string) (string, error) {
	//cryptMd5 := crypt.MD5.New()
	cryptMd5 := md5_crypt.New()
	ret, _ := cryptMd5.Generate([]byte(content), []byte("$1$"+salt))
	fmt.Println(ret)

	err := cryptMd5.Verify(ret, []byte("secret"))
	fmt.Println(err)

	// Output:
	// $5$salt$kpa26zwgX83BPSR8d7w93OIXbFt/d3UOTZaAu5vsTM6
	// <nil>
	return ret, err
}

func Bcrypt(password string) (string, error) {

	//password := "my_password"

	// 生成 hash
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("Hash:", string(hash))

	// 验证
	err = bcrypt.CompareHashAndPassword(
		hash,
		[]byte(password),
	)

	if err != nil {
		fmt.Println("Password wrong")
	} else {
		fmt.Println("Password correct")
	}
	return string(hash), nil
}

func Argon2(password string, salt string) string {

	//password := []byte("mypassword")
	//salt := []byte("randomsalt")

	hash := argon2.IDKey(
		[]byte(password),
		[]byte(salt),
		1,
		64*1024,
		4,
		32,
	)

	//fmt.Println(base64.RawStdEncoding.EncodeToString(hash))
	//return string(hash)
	return base64.RawStdEncoding.EncodeToString(hash)
}

func Md5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func Sha256(str string) string {
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}
