package wzlib

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
)

// #region ARGB

func DecodeARGB(raw []byte, width int, height int, bytesPerPixel int, formatFunc func(data []byte) color.RGBA) ([]uint8, error) {
	expectedSize := width * height * bytesPerPixel
	if len(raw) < expectedSize {
		return nil, fmt.Errorf("data too short: expected %d, got %d", expectedSize, len(raw))
	}
	pix := make([]uint8, width*height*4)
	for y := range height {
		rawOff := y * width * bytesPerPixel
		pixOff := y * width * 4
		for x := range width {
			rawIdx := rawOff + x*bytesPerPixel
			pixIdx := pixOff + x*4
			rgba := formatFunc(raw[rawIdx : rawIdx+bytesPerPixel])
			pix[pixIdx] = rgba.R
			pix[pixIdx+1] = rgba.G
			pix[pixIdx+2] = rgba.B
			pix[pixIdx+3] = rgba.A
		}
	}
	return pix, nil
}

func DecodeARGB8888(raw []byte, width int, height int) ([]uint8, error) {
	return DecodeARGB(raw, width, height, 4, func(data []byte) color.RGBA {
		// LittleEndian
		b := uint16(data[0])
		g := uint16(data[1])
		r := uint16(data[2])
		a := uint16(data[3])
		// Pre Multiplied
		if a == 0 {
			return color.RGBA{0, 0, 0, 0}
		}
		return color.RGBA{
			R: uint8(r * a / 255),
			G: uint8(g * a / 255),
			B: uint8(b * a / 255),
			A: uint8(a),
		}
	})
}

func DecodeARGB4444(raw []byte, width int, height int) ([]uint8, error) {
	return DecodeARGB(raw, width, height, 2, func(data []byte) color.RGBA {
		// LittleEndian
		val := binary.LittleEndian.Uint16(data)
		a := ((val >> 12) & 0xF) * 17
		r := ((val >> 8) & 0xF) * 17
		g := ((val >> 4) & 0xF) * 17
		b := (val & 0xF) * 17
		// Pre Multiplied
		if a == 0 {
			return color.RGBA{0, 0, 0, 0}
		}
		return color.RGBA{
			R: uint8(r * a / 255),
			G: uint8(g * a / 255),
			B: uint8(b * a / 255),
			A: uint8(a),
		}
	})
}

func DecodeARGB1555(raw []byte, width int, height int) ([]uint8, error) {
	return DecodeARGB(raw, width, height, 2, func(data []byte) color.RGBA {
		// LittleEndian
		val := binary.LittleEndian.Uint16(data)
		a := ((val >> 15) & 0x1) * 255
		r := ((val >> 10) & 0x1F) * 8
		g := ((val >> 5) & 0x1F) * 8
		b := (val & 0x1F) * 8
		// Pre Multiplied
		if a == 0 {
			return color.RGBA{0, 0, 0, 0}
		}
		return color.RGBA{
			R: uint8(r * a / 255),
			G: uint8(g * a / 255),
			B: uint8(b * a / 255),
			A: uint8(a),
		}
	})
}

func DecodeRGB565(raw []byte, width int, height int) ([]uint8, error) {
	return DecodeARGB(raw, width, height, 2, func(data []byte) color.RGBA {
		val := binary.LittleEndian.Uint16(data)
		return ParseRGB565(val)
	})
}

func ParseRGB565(val uint16) color.RGBA {
	r0 := (val >> 11) & 0x1F
	g0 := (val >> 5) & 0x3F
	b0 := val & 0x1F
	return color.RGBA{
		R: uint8((r0 << 3) | (r0 >> 2)),
		G: uint8((g0 << 2) | (g0 >> 4)),
		B: uint8((b0 << 3) | (b0 >> 2)),
		A: 0xFF,
	}
}

// #endregion

// #region DXT 3/5

// DecodeDXT is a generic DXT block decoder.
// It reconstructs 4x4 pixel blocks from compressed DXT data into raw RGBA format.
func DecodeDXT(raw []byte, width int, height int, formatFunc func(block []byte) [16]uint8) ([]uint8, error) {
	blockCountX := (width + 3) / 4
	blockCountY := (height + 3) / 4

	if len(raw) < blockCountX*blockCountY*16 {
		return nil, errors.New("raw data too short")
	}

	pix := make([]uint8, width*height*4)

	for by := range blockCountY {
		for bx := range blockCountX {
			blockOffset := (by*blockCountX + bx) * 16
			block := raw[blockOffset : blockOffset+16]

			alphas := formatFunc(block)

			val0 := binary.LittleEndian.Uint16(block[8:10])
			val1 := binary.LittleEndian.Uint16(block[10:12])
			bitcode := binary.LittleEndian.Uint32(block[12:16])

			rgba0 := ParseRGB565(val0)
			rgba1 := ParseRGB565(val1)

			var palette [4]color.RGBA
			palette[0] = rgba0
			palette[1] = rgba1

			if val0 > val1 {
				palette[2] = color.RGBA{
					R: uint8((uint16(rgba0.R)*2 + uint16(rgba1.R) + 1) / 3),
					G: uint8((uint16(rgba0.G)*2 + uint16(rgba1.G) + 1) / 3),
					B: uint8((uint16(rgba0.B)*2 + uint16(rgba1.B) + 1) / 3),
					A: 0xFF,
				}
				palette[3] = color.RGBA{
					R: uint8((uint16(rgba0.R) + uint16(rgba1.R)*2 + 1) / 3),
					G: uint8((uint16(rgba0.G) + uint16(rgba1.G)*2 + 1) / 3),
					B: uint8((uint16(rgba0.B) + uint16(rgba1.B)*2 + 1) / 3),
					A: 0xFF,
				}
			} else {
				palette[2] = color.RGBA{
					R: (rgba0.R + rgba1.R) / 2,
					G: (rgba0.G + rgba1.G) / 2,
					B: (rgba0.B + rgba1.B) / 2,
					A: 0xFF,
				}
				palette[3] = color.RGBA{0, 0, 0, 0xFF}
			}

			for j := range 4 {
				for i := range 4 {
					px := bx*4 + i
					py := by*4 + j
					if px < width && py < height {
						idx := j*4 + i
						colorIdx := (bitcode >> (idx * 2)) & 0x3
						rgba := palette[colorIdx]
						rgba.A = alphas[idx]
						if rgba.A > 0 {
							// Pre Multiplied
							rgba.R = uint8(uint16(rgba.R) * uint16(rgba.A) / 255)
							rgba.G = uint8(uint16(rgba.G) * uint16(rgba.A) / 255)
							rgba.B = uint8(uint16(rgba.B) * uint16(rgba.A) / 255)
						}
						pixIdx := (py*width + px) * 4
						pix[pixIdx] = rgba.R
						pix[pixIdx+1] = rgba.G
						pix[pixIdx+2] = rgba.B
						pix[pixIdx+3] = rgba.A
					}
				}
			}
		}
	}
	return pix, nil
}

// DecodeDXT3 wraps the generic DecodeDXT with the DXT3 alpha.
func DecodeDXT3(raw []byte, width int, height int) ([]uint8, error) {
	return DecodeDXT(raw, width, height, func(block []byte) [16]uint8 {
		var alphas [16]uint8
		for i := range 8 {
			val := block[i]
			alphas[i*2] = (val & 0x0F) * 17 // low 4-byte
			alphas[i*2+1] = (val >> 4) * 17 // high 4-byte
		}
		return alphas
	})
}

// DecodeDXT5 wraps the generic DecodeDXT with the DXT5 alpha.
func DecodeDXT5(raw []byte, width int, height int) ([]uint8, error) {
	return DecodeDXT(raw, width, height, func(block []byte) [16]uint8 {
		// Alpha endpoints: a0 and a1
		a0, a1 := block[0], block[1]
		// Create an 8-color alpha palette
		var alphaPalette [8]uint8
		alphaPalette[0], alphaPalette[1] = a0, a1
		// Interpolation mode depends on whether a0 > a1
		if a0 > a1 {
			// 6 interpolated values between a0 and a1 (1/7 steps)
			for i := 1; i < 7; i++ {
				alphaPalette[i+1] = uint8((uint16(a0)*(7-uint16(i)) + uint16(a1)*uint16(i)) / 7)
			}
		} else {
			// 4 interpolated values (1/5 steps), plus 0 and 255
			for i := 1; i < 5; i++ {
				alphaPalette[i+1] = uint8((uint16(a0)*(5-uint16(i)) + uint16(a1)*uint16(i)) / 5)
			}
			alphaPalette[6], alphaPalette[7] = 0, 255
		}
		// The alpha indices are stored in the remaining 48 bits of the first 8 bytes
		alphaIndices := binary.LittleEndian.Uint64(block[0:8]) >> 16
		var alphas [16]uint8
		for i := range 16 {
			idx := (alphaIndices >> (i * 3)) & 0x7
			alphas[i] = alphaPalette[idx]
		}
		return alphas
	})
}

// #endregion
