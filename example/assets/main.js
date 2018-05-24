console.log('Generated using Tomato!')
console.log('Visit https://github.com/ribacq/tomato')

// toggle class overlay-media on images in section on click
for (img of document.querySelectorAll('section img')) {
	img.onclick = function() {
		if (img.classList.contains('overlay-media')) {
			img.classList.remove('overlay-media')
		} else {
			img.classList.add('overlay-media')
		}
	}
}
