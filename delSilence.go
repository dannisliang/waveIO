package waveIO

import (
	"fmt"
	"math"
)

const (
	voc_Block_Len int     = 1600
	min_Voc_Eng   int     = 500
	dB            float64 = -3.0
	int16Max      int16   = 32767
)

func dBNorm(sampleBuffer []int16, sampleCount uint32) bool {
	var i uint32
	var sampleMax int16 = -int16Max

	for i = 0; i < sampleCount; i++ {
		curSample := int16(math.Abs(float64(sampleBuffer[i])))
		if curSample > sampleMax {
			sampleMax = curSample
		}
	}

	if sampleMax == 0 {
		return false
	}

	for i = 0; i < sampleCount; i++ {
		sampleBuffer[i] = int16(math.Pow(10, dB/20) * (math.Pow(2, 15) - 1) * float64(sampleBuffer[i]) / float64(sampleMax))
	}

	return true
}

func DelSilence(detFile, srcFile string) bool {
	var length, outLength uint32 = 0, 0
	var pWinBuf [voc_Block_Len + 1]int16
	var nWin, nMod, i, k, eng int
	var j, p int = 0, 0
	var old1, old2, old3, curSample int16
	pWavData := make([]int16, 0)
	pnTarget := make([]int16, 0)
	pCur := &pnTarget

	if !waveLoad(srcFile, &pWavData, &length) {
		fmt.Printf("Can not load Wave File %s !\n", srcFile)
		return false
	}
	dBNorm(pWavData, length)
	nWin = int(length) / voc_Block_Len
	nMod = int(length) % voc_Block_Len

	for i = 0; i < nWin; i++ {
		eng = 0
		for k = 0; k < voc_Block_Len; k++ {
			eng += int(math.Abs(float64(pWavData[voc_Block_Len*i+k])))
		}

		if eng > min_Voc_Eng*voc_Block_Len {
			j, p = 0, 0
			old1, old2, old3 = 0, 0, 0
			for k = 0; k < voc_Block_Len; k++ {
				curSample = pWavData[voc_Block_Len*i+k]
				if curSample == old1 && old1 == old2 && old2 == old3 {
					if p >= 0 {
						j = p
					}
				} else {
					pWinBuf[j] = curSample
					j++
					p = j - 3
				}
				old3 = old2
				old2 = old1
				old1 = curSample
			}
			for _, v := range pWinBuf[:j] {
				*pCur = append(*pCur, v)
			}
			outLength += uint32(j)
		}
	}

	eng = 0
	for i = 0; i < nMod; i++ {
		eng += int(math.Abs(float64(pWavData[voc_Block_Len*nWin+i])))
	}

	if eng > min_Voc_Eng*nMod {
		j, p = 0, 0
		old1, old2, old3 = 0, 0, 0
		for i = 0; i < nMod; i++ {
			curSample = pWavData[voc_Block_Len*nWin+i]
			if curSample == old1 && old1 == old2 && old2 == old3 {
				if p >= 0 {
					j = p
				}
			} else {
				pWinBuf[j] = curSample
				j++
				p = j - 3
			}
			old3 = old2
			old2 = old1
			old1 = curSample
		}
		for _, v := range pWinBuf[:j] {
			*pCur = append(*pCur, v)
		}
		outLength += uint32(j)
	}

	waveSave(detFile, pnTarget, outLength)
	fmt.Printf("Delete The Silence Part Of %s Successfully,And Save To %s.\n", srcFile, detFile)
	return true
}
