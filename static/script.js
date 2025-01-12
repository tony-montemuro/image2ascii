document.addEventListener('DOMContentLoaded', function() {
    /* ===== ELEMENTS ===== */
    const imageInput = document.getElementById('image');
    const imageOptions = document.getElementById('options');
    const customSize = document.getElementById('custom-size');
    const widthInput = document.getElementById('width');
    const heightInput = document.getElementById('height');
    const brightness = document.getElementById('brightness');
    const brightnessValue = document.getElementById('brightness-value');
    const uploadBtn = document.getElementById('upload');
    const error = document.getElementById('error');
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
    function addErrorMessage(message) {
        error.textContent = message;
        error.classList.remove('sr-only');
    }

    function removeErrorMessage() {
        error.textContent = '';
        error.classList.add('sr-only');
    };

    function getHeight(imageWidth, imageHeight, type) {
        return Math.round((size[type].width * imageHeight) / imageWidth / 2);
    }

    function checkImage(image) {
        const imageWidth = image.width, imageHeight = image.height;
        Object.keys(size).forEach(type => size[type].height = getHeight(imageWidth, imageHeight, type));
        imageOptions.classList.remove('sr-only');
        removeErrorMessage();
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
            imageOptions.classList.add('sr-only');
            input.value = '';
            addErrorMessage('File type not supported. Please upload a JPEG or PNG file.');
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
                customSize.classList.remove('sr-only');
            } else {
                customSize.classList.add('sr-only');
                updateSizeInputs(type);
            }
        });
    });

    uploadBtn.addEventListener('drop', event => {
        event.preventDefault();
        
        if (event.dataTransfer.files.length === 1) {
            imageInput.files = event.dataTransfer.files;
            const changeEvent = new Event('change');
            imageInput.dispatchEvent(changeEvent);
        } else {
            addErrorMessage('You can only upload one image at a time.');
        }
    });
    uploadBtn.addEventListener('dragover', event => event.preventDefault());
});