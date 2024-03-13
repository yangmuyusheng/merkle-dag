package merkledag

import (
	"hash"
)

const (
	K := 1 << 10
	BLOCK_SIZE = 256 * K
)

type Link struct {
	Name string
	Hash []byte
	Size int
}

type Object struct {
	Links []Link
	Data  []byte
}

func Add(store KVStore, node Node, h hash.Hash) []byte {
	// TODO 将分片写入到KVStore中，并返回Merkle Root
	switch node.Type() {
	case FILE:
		return StoreFile(store, node.(File), h)
	case DIR:
		return StoreDir(store, node.(Dir), h)
	default:
		return nil
	}
}

func StoreFile(store KVStore, node File, h hash.Hash) []byte {
	t := []byte("blob")
	if node.Size() > BLOCK_SIZE {
		t = []byte("list")
	}

	data := node.Bytes()
	chunks := chunkData(data)
	var merkleRoot []byte
	for _, chunk := range chunks {
		hashValue := hashData(chunk, h)
		if err := store.Put(hashValue, chunk); err != nil {
			// 处理写入KVStore失败的情况
			return nil
		}
		merkleRoot = append(merkleRoot, hashValue...)
	}

	return merkleRoot
}

func StoreDir(store KVStore, dir Dir, h hash.Hash) []byte {
	// 定义一个对象来存储文件夹内容
	obj := Object{}

	it := dir.It()
	for it.Next() {
		node := it.Node()
		switch node.Type() {
		case FILE:
			file := node.(File)
			data := file.Bytes()
			hashValue := hashData(data, h)
			if err := store.Put(hashValue, data); err != nil {
				// 处理写入KVStore失败的情况
				return nil
			}
			obj.Links = append(obj.Links, Link{Name: "file", Hash: hashValue, Size: int(file.Size())})
		case DIR:
			dir := node.(Dir)
			// 递归处理子文件夹
			childMerkleRoot := Add(store, dir, h)
			obj.Links = append(obj.Links, Link{Name: "dir", Hash: childMerkleRoot, Size: 0})
		}
	}

	// 将对象序列化为字节数组
	var serializedObj []byte
	// 这里使用你自己的序列化方法
	serializedObj = serializeObject(obj)

	// 将对象数据写入KVStore
	objHash := hashData(serializedObj, h)
	if err := store.Put(objHash, serializedObj); err != nil {
		// 处理写入KVStore失败的情况
		return nil
	}

	return objHash
}

// 将数据分片
func chunkData(data []byte) [][]byte {
	var chunks [][]byte
	const chunkSize = BLOCK_SIZE
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

// 计算数据的哈希值
func hashData(data []byte, h hash.Hash) []byte {
	h.Reset()
	h.Write(data)
	return h.Sum(nil)
}

// 自定义序列化方法
func serializeObject(obj Object) []byte {
	// 这里使用你自己的序列化方法，这里只是示例
	// 使用 encoding/gob 或 encoding/json 等进行序列化
	// 这里假设使用 encoding/gob 进行示例
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(obj)
	if err != nil {
		// 处理序列化错误
		return nil
	}
	return buf.Bytes()
}
