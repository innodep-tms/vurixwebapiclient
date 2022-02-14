package vurixwebapiclient

import (
	"bytes"
	"errors"
	"strings"
	"sync"

	"github.com/spf13/cast"
)

type MultiPartParser struct {
	mutex         *sync.Mutex
	buff          []byte
	buffSize      int
	isFirstHeader bool
	boundary      []byte
	boundarySize  int
}

func NewMultiPartParser() *MultiPartParser {
	return &MultiPartParser{
		mutex:         &sync.Mutex{},
		buff:          make([]byte, 0),
		buffSize:      0,
		isFirstHeader: false,
		boundary:      nil,
		boundarySize:  0,
	}
}

func (mp *MultiPartParser) Parse(data []byte, size int) (msg []byte, err error) {
	mp.mutex.Lock()
	if size > 0 {
		mp.buff = append(mp.buff, data[0:size]...)
		mp.buffSize += size
	}

	crlf := []byte("\r\n")
	crlfcrlf := []byte("\r\n\r\n")

	if !mp.isFirstHeader {
		// 첫 헤더를 파싱하는 부분이므로 첫줄이 HTTP프로토콜이여야 한다.
		pos := bytes.Index(mp.buff, crlf)
		firstLine := strings.Split(string(mp.buff[0:pos]), " ")
		if len(firstLine) > 2 {
			if firstLine[1] == "200" {
				sPosHeader := pos + 2
				ePosHeader := bytes.Index(mp.buff, crlfcrlf)
				if ePosHeader > 0 {
					headers := mp.parseHeader(mp.buff[sPosHeader:ePosHeader])
					contentType := mp.parseContentType(headers["Content-Type"])
					if contentType["mime"] == "multipart/mixed" {
						mp.boundary = []byte(contentType["boundary"])
						mp.boundary = append(mp.boundary, crlf...)
						mp.boundarySize = len(mp.boundary)
						//여기까지 왔으면 정상적으로 파싱된것이므로, Content-Length 길이만큼 버리고, 다시 시작한다.
						contentLength := cast.ToInt(headers["Content-Length"])
						headerEndPos := ePosHeader + 4
						bodyEndPos := headerEndPos + contentLength
						if contentLength > 0 {
							bodyEndPos += 4
						}
						if bodyEndPos <= mp.buffSize {
							mp.buff = mp.buff[bodyEndPos:]
							mp.buffSize -= bodyEndPos
							mp.isFirstHeader = true
						} else {
							// 데이터를 더 받아야 함
							msg = nil
							err = nil
						}
					} else {
						// mime타입이 틀림
						msg = nil
						err = errors.New("Invaild Mime-Type : " + headers["Content-Type"] + "\n" + string(mp.buff))
					}
				} else {
					// 데이터를 더 받아야 함
					msg = nil
					err = nil
				}
			} else {
				// 인증실패이거나 비정상 응답이므로, 에러 처리
				msg = nil
				err = errors.New("Fail to Http StatusCode : " + firstLine[1] + "\n" + string(mp.buff))
			}
		} else {
			// 메세지가 깨진것으로 판단하고 에러를 리턴
			msg = nil
			err = errors.New("Fail to HTTP Parsing : " + string(mp.buff))
		}
	} else {
		pos := bytes.Index(mp.buff, mp.boundary)
		if pos >= 0 {
			sPosHeader := pos + mp.boundarySize
			ePosHeader := bytes.Index(mp.buff, crlfcrlf)
			if ePosHeader > 0 {
				headers := mp.parseHeader(mp.buff[sPosHeader:ePosHeader])
				contentLength := cast.ToInt(headers["Content-Length"])
				headerEndPos := ePosHeader + 4
				bodyEndPos := headerEndPos + contentLength
				if contentLength > 0 {
					bodyEndPos += 4
				}
				if bodyEndPos <= mp.buffSize {
					// 메세지 파싱완료
					msg = mp.buff[headerEndPos : headerEndPos+contentLength]
					err = nil

					// 파싱완료된 메모리 제거
					mp.buff = mp.buff[bodyEndPos:]
					mp.buffSize -= bodyEndPos

				} else {
					// 데이터를 더 받아야 함
					msg = nil
					err = nil
				}
			} else {
				// 데이터를 더 받아야 함
				msg = nil
				err = nil
			}
		} else {
			// 데이터를 더 받아야 함
			msg = nil
			err = nil
		}
	}
	mp.mutex.Unlock()
	return
}
func (mp *MultiPartParser) parseHeader(data []byte) map[string]string {
	sHeader := strings.Split(string(data), "\r\n")
	headers := make(map[string]string)
	for i := range sHeader {
		sHeader := strings.SplitN(sHeader[i], ":", 2)
		if len(sHeader) == 2 {
			sHeader[0] = strings.TrimSpace(sHeader[0])
			sHeader[1] = strings.TrimSpace(sHeader[1])
			headers[sHeader[0]] = sHeader[1]
		}
	}
	return headers
}

func (mp *MultiPartParser) parseContentType(data string) map[string]string {
	sData := strings.Split(strings.ToLower(data), ";")
	contentType := make(map[string]string)
	for i := range sData {
		if strings.Contains(sData[i], "/") {
			contentType["mime"] = strings.TrimSpace(sData[i])
		} else if strings.Contains(sData[i], "=") {
			keyvalue := strings.SplitN(sData[i], "=", 2)
			keyvalue[0] = strings.TrimSpace(keyvalue[0])
			keyvalue[1] = strings.TrimSpace(keyvalue[1])
			contentType[keyvalue[0]] = keyvalue[1]
		}
	}
	return contentType
}
