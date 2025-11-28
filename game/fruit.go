package game

import "image/color"

type Fruit struct {
	X         int
	Y         int
	Type      FruitType
	Active    bool
	Eaten     bool
	SpawnTime int
}

func NewFruit(level int) *Fruit {
	// Fruit type based on level (classic Pac-Man progression)
	fruitType := FruitType(level - 1)
	if fruitType > FruitKey {
		fruitType = FruitKey
	}

	return &Fruit{
		X:      14,
		Y:      17,
		Type:   fruitType,
		Active: true,
		Eaten:  false,
	}
}

func (f *Fruit) GetColor() color.RGBA {
	switch f.Type {
	case FruitCherry:
		return color.RGBA{255, 0, 0, 255} // Red
	case FruitStrawberry:
		return color.RGBA{255, 100, 100, 255} // Light red
	case FruitOrange:
		return color.RGBA{255, 165, 0, 255} // Orange
	case FruitApple:
		return color.RGBA{255, 0, 0, 255} // Red
	case FruitMelon:
		return color.RGBA{0, 255, 0, 255} // Green
	case FruitGalaxian:
		return color.RGBA{0, 255, 255, 255} // Cyan
	case FruitBell:
		return color.RGBA{255, 215, 0, 255} // Gold
	case FruitKey:
		return color.RGBA{255, 255, 0, 255} // Yellow
	default:
		return color.RGBA{255, 0, 0, 255}
	}
}

func (f *Fruit) GetSymbol() string {
	switch f.Type {
	case FruitCherry:
		return "🍒"
	case FruitStrawberry:
		return "🍓"
	case FruitOrange:
		return "🍊"
	case FruitApple:
		return "🍎"
	case FruitMelon:
		return "🍉"
	case FruitGalaxian:
		return "🚀"
	case FruitBell:
		return "🔔"
	case FruitKey:
		return "🔑"
	default:
		return "🍒"
	}
}
