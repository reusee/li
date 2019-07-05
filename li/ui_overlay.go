package li

func OverlayUI(
	overlays []Overlay,
	views Views,
) (ret []Element) {

	for _, overlay := range overlays {
		if overlay.Element == nil {
			continue
		}
		ret = append(ret, overlay.Element)
	}

	return
}
