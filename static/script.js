document.addEventListener('DOMContentLoaded', function() {
    /* ===== ELEMENTS ===== */
    const form = this.getElementById('form');
    const imageInput = this.getElementById('image');
    const imageOptions = this.getElementById('options');
    const customSize = this.getElementById('custom-size');
    const widthInput = this.getElementById('width');
    const heightInput = this.getElementById('height');
    const brightness = this.getElementById('brightness');
    const brightnessValue = this.getElementById('brightness-value');
    const uploadBtn = this.getElementById('upload');
    const error = this.getElementById('error');
    const imagePlaceholder = this.getElementById('img-placeholder');
    const thumbnailWrapper = this.getElementById('thumbnail-wrapper');
    const thumbnail = this.getElementById('thumbnail');
    const thumbnailName = this.getElementById('thumbnail-name');
    const sizeContainer = this.getElementById('size');
    const output = this.getElementById('output');
    const copySuccess = this.getElementById("output-copy-success");
    const copyError = this.getElementById("output-copy-failure");
    const outputContainer = this.getElementById("output-wrapper");
    const sizeRadios = sizeContainer.querySelectorAll('input[name="size"]');
    const sizeRadioLabels = sizeContainer.getElementsByTagName('label'); 

    /* ===== VARIABLES ===== */
    const size = {
        twitch: {
            width: 30,
            height: undefined,
            maxHeight: 16
        },
        discord: {
            width: 32,
            height: undefined,
            maxHeight: 62
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

    function popIn(element) {
        show(element);
        element.classList.add("animate-popin");
    }

    function popOut(element) {
        hide(element);
        element.classList.remove("animate-popin");
    }

    function addErrorMessage(message) {
        show(error);
        error.textContent = message;
    };

    function removeErrorMessage() {
        hide(error);
        error.textContent = '';
    }

    function hideOptions(message) {
        thumbnail.src = '';
        thumbnail.alt = '';
        thumbnailName.textContent = '';
        hide(thumbnailWrapper);
        show(imagePlaceholder);

        imageInput.value = '';
        hide(imageOptions);
        addErrorMessage(message);
    }

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
        const maxHeight = size[type].maxHeight ?? Number.MAX_SAFE_INTEGER;
        const calculatedHeight = Math.round((size[type].width * imageHeight) / imageWidth / 2);
        return Math.min(calculatedHeight, maxHeight);
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
            hideOptions('File type not supported. Please upload a JPEG or PNG file.');
        }
    };

    function updateSizeInputs(type) {
        widthInput.value = size[type].width;
        heightInput.value = size[type].height;
    };


    /* ===== EVENT LISTENERS ===== */

    // Upload button
    uploadBtn.addEventListener('keydown', event => event.key === "Enter" ? imageInput.click() : null);
    uploadBtn.addEventListener('drop', event => {
        event.preventDefault();
        
        if (event.dataTransfer.files.length === 1) {
            imageInput.files = event.dataTransfer.files;
            const changeEvent = new Event('change');
            imageInput.dispatchEvent(changeEvent);
        } else {
            hideOptions('You can only upload one image at a time.');
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
    for (const label of sizeRadioLabels) {
        label.addEventListener('keydown', event => {
            if (event.key === " ") {
                event.preventDefault();
                event.target.click();
                event.target.focus();
            }
        });
    }

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

            output.textContent = '';
            if (response.status !== 200) {
                throw new Error(data.error);
            }

            removeErrorMessage();
            data.forEach(row => {
                output.textContent += row + "\n";
            });
            show(outputContainer);

        } catch(error) {
            addErrorMessage(error.message);
        }
    });

    // Output
    let timeout;
    output.addEventListener('click', async event => {
        clearTimeout(timeout);

        const element = event.target;
        const text = element.textContent;
        const type = "text/plain";
        const blob = new Blob([text], {type});
        const data = [new ClipboardItem({[type]: blob})];

        try {
            await navigator.clipboard.write(data);
            popIn(copySuccess);
            timeout = setTimeout(() => popOut(copySuccess), 1500);
        } catch (error) {
            popIn(copyError);
            timeout = setTimeout(() => popOut(copyError), 1500);
        }
    });
});