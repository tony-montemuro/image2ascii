document.addEventListener('DOMContentLoaded', function() {
    /* ===== ELEMENTS ===== */
    const imageInput = document.getElementById('image');
    const imageOptions = document.getElementById('options');
    const customSize = document.getElementById('custom-size');
    const twitchBtn = document.getElementById('twitch');
    const discordBtn = document.getElementById('discord');
    const widthInput = document.getElementById('width');
    const heightInput = document.getElementById('height');
    const brightness = document.getElementById('brightness');
    const brightnessValue = document.getElementById('brightness-value');
    const sizeRadios = document.getElementsByName('size');

    /* ===== VARIABLES ===== */
    const size = {
        twitch: {
            width: 30,
            height: undefined
        },
        discord: {
            width: 33,
            height: undefined
        },
        small: {
            width: 20,
            height: undefined
        },
        medium: {
            width: 50,
            height: undefined
        },
        large: {
            width: 100,
            height: undefined
        }
    };

    /* ===== FUNCTIONS ===== */
    function getHeight(imageWidth, imageHeight, type) {
        return Math.round((size[type].width * imageHeight) / imageWidth / 2);
    }

    function checkImage(image) {
        const imageWidth = image.width, imageHeight = image.height;
        Object.keys(size).forEach(type => size[type].height = getHeight(imageWidth, imageHeight, type));
        imageOptions.classList.remove('opacity-0');
    };

    function handleNewImage(input, files) {
        const img = files[0];
        const validTypes = ['image/jpeg', 'image/png'];

        if (validTypes.includes(img.type)) {
            const image = new Image();
            image.src = URL.createObjectURL(img);
            image.onload = function() {
                checkImage(this);
            }
        } else {
            imageOptions.classList.add('opacity-0');
            input.value = '';
        }
    };

    function updateSizeInputs(type) {
        widthInput.value = size[type].width;
        heightInput.value = size[type].height;
    };

    /* ===== EVENT LISTENERS ===== */
    imageInput.addEventListener('change', event => {
        const input = event.target;
        const files = input.files;

        if (files.length > 0) {
            handleNewImage(input, files);
        }
    });

    brightness.addEventListener('input', event => {
        brightnessValue.value = event.target.value;
    });
    brightnessValue.addEventListener('change', event => {
        brightness.value = event.target.value;
    });

    sizeRadios.forEach(radio => {
        radio.addEventListener('click', event => {
            const type = event.target.value;

            if (type === "custom") {
                customSize.classList.remove('sr-only', 'w-0', 'h-0', 'absolute');
            } else {
                customSize.classList.add('sr-only', 'w-0', 'h-0', 'absolute');
                updateSizeInputs(type);
            }
        });
    });
});