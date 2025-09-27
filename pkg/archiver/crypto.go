package archiver

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

const (
	saltSize   = 32        // Salt 的大小（字节）
	keySize    = 32        // AES-256 需要 32 字节的密钥
	pbkdf2Iter = 600000    // PBKDF2 迭代次数
	chunkSize  = 64 * 1024 // 每次加密/解密的块大小 (64KB)
)

// --- 加密实现 ---

// gcmStreamEncrypter 实现了 io.WriteCloser 接口，用于流式加密
type gcmStreamEncrypter struct {
	w         io.Writer   // 底层写入器
	aead      cipher.AEAD // AEAD 加密器
	nonce     []byte      // 当前的 Nonce
	buffer    []byte      // 暂存待加密的数据
	chunkSize int
}

// encryptStream 返回一个加密写入器
func encryptStream(w io.Writer, password string) (io.WriteCloser, error) {
	// 1. 生成 Salt 和起始 Nonce
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("生成 salt 失败: %w", err)
	}
	// GCM 的 Nonce 大小通常是 12 字节
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("生成 nonce 失败: %w", err)
	}

	// 2. 派生密钥
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iter, keySize, sha3.New512)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 3. 写入 Salt 和 Nonce 头信息
	if _, err := w.Write(salt); err != nil {
		return nil, err
	}
	if _, err := w.Write(nonce); err != nil {
		return nil, err
	}

	return &gcmStreamEncrypter{
		w:         w,
		aead:      gcm,
		nonce:     nonce,
		buffer:    make([]byte, 0, chunkSize),
		chunkSize: chunkSize,
	}, nil
}

// Write 实现了 io.Writer 接口
func (e *gcmStreamEncrypter) Write(p []byte) (n int, err error) {
	e.buffer = append(e.buffer, p...)
	// 当缓冲区满时，加密并写入一个块
	for len(e.buffer) >= e.chunkSize {
		if err := e.encryptChunk(e.buffer[:e.chunkSize]); err != nil {
			return 0, err
		}
		e.buffer = e.buffer[e.chunkSize:]
	}
	return len(p), nil
}

// Close 实现了 io.Closer 接口
func (e *gcmStreamEncrypter) Close() error {
	// 加密并写入缓冲区中剩余的最后一个数据块
	if len(e.buffer) > 0 {
		if err := e.encryptChunk(e.buffer); err != nil {
			return err
		}
	}
	// 写入一个零长度的块作为结束标记 (EOF)
	return binary.Write(e.w, binary.LittleEndian, uint32(0))
}

// encryptChunk 加密单个数据块并写入底层 writer
func (e *gcmStreamEncrypter) encryptChunk(plaintext []byte) error {
	ciphertext := e.aead.Seal(nil, e.nonce, plaintext, nil)
	// 写入密文长度 (uint32)
	if err := binary.Write(e.w, binary.LittleEndian, uint32(len(ciphertext))); err != nil {
		return err
	}
	// 写入密文
	if _, err := e.w.Write(ciphertext); err != nil {
		return err
	}
	// 递增 Nonce，防止重用 (对 12 字节的大数进行递增)
	for i := len(e.nonce) - 1; i >= 0; i-- {
		e.nonce[i]++
		if e.nonce[i] != 0 {
			break
		}
	}
	return nil
}

// --- 解密实现 ---

// gcmStreamDecrypter 实现了 io.Reader 接口
type gcmStreamDecrypter struct {
	r      io.Reader
	aead   cipher.AEAD
	nonce  []byte
	buffer []byte // 暂存已解密但尚未被读取的数据
	eof    bool
}

// decryptStream 返回一个解密读取器
func decryptStream(r io.Reader, password string) (io.Reader, error) {
	// 1. 读取 Salt 和起始 Nonce
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(r, salt); err != nil {
		return nil, fmt.Errorf("读取 salt 失败: %w", err)
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(r, nonce); err != nil {
		return nil, fmt.Errorf("读取 nonce 失败: %w", err)
	}

	// 2. 派生密钥并创建 AEAD
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iter, keySize, sha3.New512)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &gcmStreamDecrypter{
		r:     r,
		aead:  gcm,
		nonce: nonce,
	}, nil
}

// Read 实现了 io.Reader 接口
func (d *gcmStreamDecrypter) Read(p []byte) (n int, err error) {
	// 如果上次解密的数据还没读完，先从 buffer 中读取
	if len(d.buffer) > 0 {
		n = copy(p, d.buffer)
		d.buffer = d.buffer[n:]
		return n, nil
	}
	// 如果已经读到流的末尾，返回 EOF
	if d.eof {
		return 0, io.EOF
	}

	// 从底层 reader 中读取并解密下一个块
	if err := d.decryptChunk(); err != nil {
		return 0, err
	}
	// 再次调用 Read 来从新填充的 buffer 中读取数据
	return d.Read(p)
}

// decryptChunk 解密单个数据块并放入 buffer
func (d *gcmStreamDecrypter) decryptChunk() error {
	var chunkSize uint32
	// 读取密文长度
	if err := binary.Read(d.r, binary.LittleEndian, &chunkSize); err != nil {
		return err
	}

	// 零长度块是结束标记
	if chunkSize == 0 {
		d.eof = true
		return nil
	}

	// 读取密文
	ciphertext := make([]byte, chunkSize)
	if _, err := io.ReadFull(d.r, ciphertext); err != nil {
		return err
	}

	// 解密
	plaintext, err := d.aead.Open(nil, d.nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("解密块失败 (可能密码错误或数据损坏): %w", err)
	}
	d.buffer = plaintext

	// 递增 Nonce
	for i := len(d.nonce) - 1; i >= 0; i-- {
		d.nonce[i]++
		if d.nonce[i] != 0 {
			break
		}
	}
	return nil
}
