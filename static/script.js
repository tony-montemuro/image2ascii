document.addEventListener('DOMContentLoaded', function() {
    /* ===== ELEMENTS ===== */
    const form = this.getElementById('form');
    const imageInput = document.getElementById('image');
    const imageOptions = document.getElementById('options');
    const customSize = document.getElementById('custom-size');
    const widthInput = document.getElementById('width');
    const heightInput = document.getElementById('height');
    const brightness = document.getElementById('brightness');
    const brightnessValue = document.getElementById('brightness-value');
    const uploadBtn = document.getElementById('upload');
    const error = document.getElementById('error');
    const imagePlaceholder = document.getElementById('img-placeholder');
    const thumbnailWrapper = document.getElementById('thumbnail-wrapper');
    const thumbnail = document.getElementById('thumbnail');
    const thumbnailName = document.getElementById('thumbnail-name');
    const sizeContainer = document.getElementById('size');
    const apiError = document.getElementById('api-error');
    const output = document.getElementById('output');
    const sizeRadios = sizeContainer.querySelectorAll('input[name="size"]');

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
    function show(element) {
        element.classList.remove('sr-only');
    }

    function hide(element) {
        element.classList.add('sr-only');
    }

    function addErrorMessage(message) {
        thumbnail.src = '';
        thumbnail.alt = '';
        thumbnailName.textContent = '';
        hide(thumbnailWrapper);
        show(imagePlaceholder);

        imageInput.value = '';
        hide(imageOptions);
        show(error);
        error.textContent = message;
    };

    function displayOptions(image) {
        thumbnail.src = image.src;
        thumbnail.alt = 'Your image';
        thumbnailName.textContent = image.name;
        show(thumbnailWrapper);
        hide(imagePlaceholder);

        show(imageOptions);
        hide(error);
        error.textContent = '';

        updateSizeInputs(sizeContainer.querySelector('input:checked').value);
    };

    function getHeight(imageWidth, imageHeight, type) {
        return Math.round((size[type].width * imageHeight) / imageWidth / 2);
    }

    function checkImage(image) {
        const imageWidth = image.width, imageHeight = image.height;
        Object.keys(size).forEach(type => size[type].height = getHeight(imageWidth, imageHeight, type));
        displayOptions(image);
    };

    function handleNewImage(files) {
        const img = files[0];
        const validTypes = ['image/jpeg', 'image/png'];

        if (validTypes.includes(img.type)) {
            const image = new Image();
            image.src = URL.createObjectURL(img);
            image.onload = function() {
                image.setAttribute('name', img.name);
                checkImage(this);
            }
        } else {
            addErrorMessage('File type not supported. Please upload a JPEG or PNG file.');
        }
    };

    function updateSizeInputs(type) {
        widthInput.value = size[type].width;
        heightInput.value = size[type].height;
    };

    function addApiErrorMessage(message) {
        apiError.textContent = message;
        show(apiError);
    }

    function removeApiErrorMessage() {
        apiError.textContent = '';
        hide(apiError);
    }

    /* ===== EVENT LISTENERS ===== */

    // Upload button
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
    imageInput.addEventListener('change', event => {
        const input = event.target;
        const files = input.files;

        if (files.length > 0) {
            handleNewImage(files);
        }
    });

    // Size radio buttons
    sizeRadios.forEach(radio => {
        const sizeInputs = customSize.getElementsByTagName('input');
        
        radio.addEventListener('click', event => {
            const changeUsability = isReadOnly => {
                for (const input of sizeInputs) {
                    if (isReadOnly) {
                        input.removeAttribute('readonly');
                        input.classList.remove('bg-gray-100');
                    } else {
                        input.setAttribute('readonly', 'readonly');
                        input.classList.add('bg-gray-100');
                    }
                }
            }

            const type = event.target.value;
            if (type === "custom") {
                changeUsability(true);
            } else {
                changeUsability(false);
                updateSizeInputs(type);
            }
        });
    });

    // Brightness input
    brightness.addEventListener('input', event => {
        brightnessValue.value = event.target.value;
    });
    brightnessValue.addEventListener('change', event => {
        brightness.value = event.target.value;
    });

    // Form submission
    form.addEventListener('submit', async event => {
        event.preventDefault();

        const form = event.target;
        const action = form.action;
        const method = form.method;
        const formData = new FormData(form);
        formData.delete('size');

        try {
            let response = await fetch(action, {
                method,
                body: formData
            });
            let data = await response.json();

            output.value = '';
            if (response.status !== 200) {
                throw new Error(data.error);
            }

            removeApiErrorMessage();
            data.forEach(row => {
                output.value += row + "\n";
            });

            console.log(data);
        } catch(error) {
            addApiErrorMessage(error.message);
        }
    });
});