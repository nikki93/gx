package main

//
// Built-ins
//

//gxsl:extern vec2
type Vec2 struct {
	X, Y float64
}

//gxsl:extern vec4
type Vec4 struct {
	X, Y, Z, W float64
}

//gxsl:extern +
func (v Vec4) Add(u Vec4) Vec4

//gxsl:extern -
func (v Vec4) Negate() Vec4

//gxsl:extern *
func (v Vec4) Multiply(u Vec4) Vec4

//gxsl:extern *
func (v Vec4) Scale(f float64) Vec4

//gxsl:extern dot
func (v Vec4) DotProduct(u Vec4) float64

//gxsl:extern sampler2D
type Sampler2D struct{}

//gxsl:extern texture2D
func Texture2D(sampler Sampler2D, coord Vec2) Vec4

//gxsl:extern gl_FragColor
var gl_FragColor Vec4

//
// Shader
//

//gx:extern INVALID
type Varyings struct {
	FragTexCoord Vec2
	FragColor    Vec4
}

//gx:extern INVALID
type RedTextureParams struct {
	ColDiffuse Vec4
	Texture0   Sampler2D
	Triple     FloatTriple
	Tricky     Varyings
}

//gx:extern INVALID
func scaleByFive(vec Vec4) Vec4 {
	return scaleByNum(vec, 3).Add(scaleByTwo(vec))
}

//gx:extern INVALID
func scaleByTwo(vec Vec4) Vec4 {
	return scaleByNum(vec, 2)
}

//gx:extern INVALID
func scaleByNum(vec Vec4, num float64) Vec4 {
	return vec.Scale(num)
}

//gx:extern INVALID
type FloatPair struct {
	A, B float64
}

//gx:extern INVALID
type FloatTriple struct {
	A, B, C float64
}

//gx:extern INVALID
var red = Vec4{-1, -0.2, -0.2, -1}.Negate()

//gxsl:shader
func redTextureShader(uniforms RedTextureParams, varyings Varyings) {
	result := red

	texelColor := Texture2D(uniforms.Texture0, varyings.FragTexCoord)
	result = result.Multiply(texelColor)

	result = result.Multiply(uniforms.ColDiffuse)

	result = result.Multiply(varyings.FragColor)

	result = scaleByFive(result.Scale(result.DotProduct(Vec4{1, 0, 0, 1})))

	floatPair := FloatPair{2, 3}
	result = result.Scale(floatPair.B)

	gl_FragColor = result
}

//
// Main
//

func main() {
}
