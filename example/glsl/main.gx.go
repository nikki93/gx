package main

//
// Built-ins
//

type Vec2 struct {
	X, Y float64
}

func (v Vec2) Multiply(u Vec2) Vec2

type Vec4 struct {
	X, Y, Z, W float64
}

func (v Vec4) Multiply(u Vec4) Vec4

type Sampler2D struct{}

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

//gx:glsl
func redTextureShader(uniforms RedTextureParams, varyings Varyings) {
	result := Vec4{1, 0.2, 0.2, 1}

	texelColor := Texture2D(uniforms.Texture0, varyings.FragTexCoord)
	result = result.Multiply(texelColor)

	result = result.Multiply(uniforms.ColDiffuse)

	result = result.Multiply(varyings.FragColor)

	gl_FragColor = result
}

//
// GX API
//

//gx:extern INVALID
func GLSL(shaderFunc interface{}) (result string) { return }

//
// Main
//

func main() {
	println(GLSL(redTextureShader))
}
