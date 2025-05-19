package textures

import (
	"encoding/json"
	"image"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var Base_Shader = `//kage:unit pixels
			package main

			func Fragment(targetCoords vec4, srcPos vec2, _ vec4) vec4 {
				col := imageSrc0At(srcPos.xy)
				return vec4(col.x, col.y, col.z, col.w)
			}
`

type RenderableTexture interface {
	Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions)
	Update()
	GetTexture() *ebiten.Image
	RefreshTexture()
	SetUniforms(uniforms map[string]any)
}

type Texture struct {
	Path     string
	Img      *ebiten.Image
	Shader   *ebiten.Shader
	Uniforms map[string]any
}

func NewTexture(img_path string, shader string) *Texture {
	texture := Texture{}

	texture.Path = img_path

	timg, _, err := ebitenutil.NewImageFromFile(img_path)
	if err != nil {
		panic(err)
	}
	texture.Img = timg

	if shader == "" {
		shader = Base_Shader
	}

	texture.Shader, err = ebiten.NewShader([]byte(shader))
	if err != nil {
		panic(err)
	}

	return &texture
}

func (texture *Texture) Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = texture.Img
	opts.Uniforms = texture.Uniforms
	opts.GeoM = op.GeoM
	screen.DrawRectShader(texture.Img.Bounds().Dx(), texture.Img.Bounds().Dy(), texture.Shader, opts)
}

func (texture *Texture) SetUniforms(uniforms map[string]any) {
	texture.Uniforms = uniforms
}

func (texture *Texture) GetTexture() *ebiten.Image {
	return texture.Img
}

func (texture *Texture) Update() {}

func (texture *Texture) RefreshTexture() {
	texture.Path = texture.Path

	timg, _, err := ebitenutil.NewImageFromFile(texture.Path)
	if err != nil {
		panic(err)
	}
	texture.Img = timg
}

type Animation struct {
	Frames             []*ebiten.Image
	Animation_Progress int
	Speed              float64
	Timer              float64
}

type SpriteSheetData struct {
	Frames [][][]int
	Speed  float64
}

type AnimatedTexture struct {
	Path              string
	Animations        []Animation
	Modified          bool
	Shader            *ebiten.Shader
	Uniforms          map[string]any
	Current_Animation int
}

func NewAnimatedTexture(path string, shader string) *AnimatedTexture {
	sprite_sheet, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		panic(err)
	}

	animated_texture := AnimatedTexture{}

	animated_texture.Path = path

	temp := SpriteSheetData{}

	json_path := strings.Replace(path, "png", "json", -1)

	temp_data, _ := os.ReadFile(json_path)

	json.Unmarshal(temp_data, &temp)

	animations := []Animation{}

	for animation := 0; animation < len(temp.Frames); animation++ {
		animations = append(animations, Animation{})
		animations[animation].Speed = float64(temp.Speed)
		animations[animation].Timer = 0
		for fram := 0; fram < len(temp.Frames[animation]); fram++ {
			frame := []float64{float64(int(temp.Frames[animation][fram][0])), float64(int(temp.Frames[animation][fram][1])), float64(int(temp.Frames[animation][fram][2])), float64(int(temp.Frames[animation][fram][3]))}
			animations[animation].Frames = append(animations[animation].Frames, ebiten.NewImageFromImage(sprite_sheet.SubImage(image.Rect(int(frame[0]), int(frame[1]), int(frame[2]), int(frame[3])))))
		}
	}
	animated_texture.Animations = animations

	if shader == "" {
		shader = Base_Shader
	}

	animated_texture.Shader, err = ebiten.NewShader([]byte(shader))
	if err != nil {
		panic(err)
	}

	return &animated_texture
}

func (texture *AnimatedTexture) Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = texture.Animations[texture.Current_Animation].Frames[texture.Animations[texture.Current_Animation].Animation_Progress]
	opts.Uniforms = texture.Uniforms
	opts.GeoM = op.GeoM
	screen.DrawRectShader(texture.Animations[texture.Current_Animation].Frames[texture.Animations[texture.Current_Animation].Animation_Progress].Bounds().Dx(), texture.Animations[texture.Current_Animation].Frames[texture.Animations[texture.Current_Animation].Animation_Progress].Bounds().Dy(), texture.Shader, opts)
}

func (texture *AnimatedTexture) SetUniforms(uniforms map[string]any) {
	texture.Uniforms = uniforms
}

func (texture *AnimatedTexture) RefreshTexture() {
	sprite_sheet, _, err := ebitenutil.NewImageFromFile(texture.Path)
	if err != nil {
		panic(err)
	}

	temp := SpriteSheetData{}

	json_path := strings.Replace(texture.Path, "png", "json", -1)

	temp_data, _ := os.ReadFile(json_path)

	json.Unmarshal(temp_data, &temp)

	animations := []Animation{}

	for anim := 0; anim < len(temp.Frames); anim++ {
		animations = append(animations, Animation{})
		animations[anim].Speed = float64(temp.Speed)
		animations[anim].Timer = texture.Animations[anim].Timer
		animations[anim].Animation_Progress = texture.Animations[anim].Animation_Progress
		for fram := 0; fram < len(temp.Frames[anim]); fram++ {
			frame := []float64{float64(int(temp.Frames[anim][fram][0])), float64(int(temp.Frames[anim][fram][1])), float64(int(temp.Frames[anim][fram][2])), float64(int(temp.Frames[anim][fram][3]))}
			animations[anim].Frames = append(animations[anim].Frames, ebiten.NewImageFromImage(sprite_sheet.SubImage(image.Rect(int(frame[0]), int(frame[1]), int(frame[2]), int(frame[3])))))
		}
	}
	texture.Animations = animations
}

func (texture *AnimatedTexture) Update() {
	texture.Animations[texture.Current_Animation].Timer -= texture.Animations[texture.Current_Animation].Speed

	if texture.Animations[texture.Current_Animation].Timer < 0 {
		texture.Animations[texture.Current_Animation].Animation_Progress += 1
		if texture.Animations[texture.Current_Animation].Animation_Progress >= len(texture.Animations[texture.Current_Animation].Frames) {
			texture.Animations[texture.Current_Animation].Animation_Progress = 0
		}
		texture.Animations[texture.Current_Animation].Timer = 1
	}
}

func (texture *AnimatedTexture) GetTexture() *ebiten.Image {
	return texture.Animations[texture.Current_Animation].Frames[texture.Animations[texture.Current_Animation].Animation_Progress]
}
