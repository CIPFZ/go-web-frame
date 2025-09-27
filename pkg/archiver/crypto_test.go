package archiver

import (
	"bytes"
	"io"
	"testing"
)

func TestCryptoStreamCycle(t *testing.T) {
	password := "my-strong-password-123"
	plaintext := []byte("this is a super secret message that needs to be encrypted")

	// 加密
	var buf bytes.Buffer
	encWriter, err := encryptStream(&buf, password)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	_, err = encWriter.Write(plaintext)
	if err != nil {
		t.Fatalf("写入加密流失败: %v", err)
	}
	// 关闭是必须的，它会写入 GCM 的认证标签
	if err := encWriter.Close(); err != nil {
		t.Fatalf("关闭加密流失败: %v", err)
	}

	// 使用正确密码解密
	decReader, err := decryptStream(&buf, password)
	if err != nil {
		t.Fatalf("使用正确密码解密失败: %v", err)
	}
	decrypted, err := io.ReadAll(decReader)
	if err != nil {
		t.Fatalf("读取解密流失败: %v", err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("解密后的内容不匹配. 期望: %s, 得到: %s", plaintext, decrypted)
	}

	// 使用错误密码解密
	// 需要重新使用加密后的数据，因为读取器会消耗它
	encryptedData := buf.Bytes()
	wrongPassReader, err := decryptStream(bytes.NewReader(encryptedData), "wrong-password")
	if err != nil {
		t.Fatalf("使用错误密码创建解密器失败: %v", err)
	}
	_, err = io.ReadAll(wrongPassReader)
	// AES-GCM 会在读取时因为认证标签不匹配而报错
	if err == nil {
		t.Error("期望使用错误密码解密时发生错误，但没有发生")
	}
}
