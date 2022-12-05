package main

//
// GLSL API
//

//gx:extern INVALID
func GLSL(shaderFunc interface{}) (result string) { return }

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

var gl_FragColor Vec4

//
// Shader
//

type Varyings struct {
	FragTexCoord Vec2
	FragColor    Vec4
}

type RedTextureParams struct {
	ColDiffuse Vec4
	Texture0   Sampler2D
}

//gxsl:entry
func redTextureShader(uniforms RedTextureParams, varyings Varyings) {
	result := Vec4{-1, -0.2, -0.2, -1}.Negate()

	texelColor := Texture2D(uniforms.Texture0, varyings.FragTexCoord)
	result = result.Multiply(texelColor)

	result = result.Multiply(uniforms.ColDiffuse)

	result = result.Multiply(varyings.FragColor)

	result = result.Scale(result.DotProduct(Vec4{1, 0, 0, 1}))

	gl_FragColor = result
}

//
// Main
//

func main() {
	println(GLSL(redTextureShader))
}
