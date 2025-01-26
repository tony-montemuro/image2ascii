document.addEventListener('DOMContentLoaded', function() {
    /* ===== ELEMENTS ===== */
    const form = this.getElementById('form');
    const imageInput = this.getElementById('image');
    const imageOptions = this.getElementById('options');
    const customSize = this.getElementById('custom-size');
    const widthInput = this.getElementById('width');
    const heightInput = this.getElementById('height');
    const exposure = this.getElementById('exposure');
    const exposureValue = this.getElementById('exposure-value');
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
    const submitBtn = this.getElementById("submit");
    const submitBtnText = this.getElementById("submit-btn-text");
    const submitBtnSpinner = this.getElementById("submit-btn-spinner");
    const sizeRadios = sizeContainer.querySelectorAll('input[name="size"]');
    const sizeRadioLabels = sizeContainer.getElementsByTagName('label'); 
    const widthAndHeightInputs = customSize.getElementsByTagName('input');

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
            width: 30,
            height: undefined
        },
        medium: {
            width: 90,
            height: undefined
        },
        large: {
            width: 150,
            height: undefined
        }
    };
    let clipboardModalTimeout;

    /* ===== FUNCTIONS ===== */
    function show(element) {
        element.classList.remove('sr-only');
    }

    function hide(element) {
        element.classList.add('sr-only');
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

    function updateWidthAndHeight(type) {
        widthInput.value = size[type].width;
        heightInput.value = size[type].height;
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

        updateWidthAndHeight(sizeContainer.querySelector('input:checked').value);
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

    function setSubmitting(isSubmitting) {
        const submittingClasses = ['cursor-not-allowed', 'bg-blue-500/90'];
        const normalClasses = ['hover:bg-blue-500/90'];

        if (isSubmitting) {
            hide(submitBtnText);
            show(submitBtnSpinner);
            submitBtn.disabled = true;
            submitBtn.classList.add(...submittingClasses);
            submitBtn.classList.remove(...normalClasses);
        } else {
            hide(submitBtnSpinner);
            show(submitBtnText);
            submitBtn.removeAttribute("disabled");
            submitBtn.classList.remove(...submittingClasses);
            submitBtn.classList.add(...normalClasses);
        }
    }

    // actions

    async function getOutput(form) {
        const action = form.action;
        const method = form.method;
        const formData = new FormData(form);
        formData.delete('size');

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
    }

    function uploadBtnDropAction(event) {
        event.preventDefault();
        
        if (event.dataTransfer.files.length === 1) {
            imageInput.files = event.dataTransfer.files;
            const changeEvent = new Event('change');
            imageInput.dispatchEvent(changeEvent);
        } else {
            hideOptions('You can only upload one image at a time.');
        }
    }

    function imageInputChangeAction(event) {
        const input = event.target;
        const files = input.files;

        if (files.length > 0) {
            handleNewImage(files);
        }
    }

    function sizeRadioClickAction(event) {
        const changeUsability = enabling => {
            for (const input of widthAndHeightInputs) {
                if (enabling) {
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
            updateWidthAndHeight(type);
        }
    }

    function sizeRadioLabelKeydownAction(event) {
        if (event.key === " ") {
            event.preventDefault();
            event.target.click();
            event.target.focus();
        }
    }

    async function formSubmitAction(event) {
        event.preventDefault();
        const form = event.target;

        setSubmitting(true);
        try {
            getOutput(form);
        } catch(error) {
            addErrorMessage(error.message);
        } finally {
            setSubmitting(false);
        }
    }

    async function outputClickAction(event) {
        const popIn = element => {
            show(element);
            element.classList.add('animate-popin');
        }

        const popOut = element => {
            element.classList.remove('animate-popin');
            element.classList.add('animate-popout');
        }

        clearTimeout(clipboardModalTimeout);

        const element = event.target;
        const text = element.textContent;
        const type = "text/plain";
        const blob = new Blob([text], {type});
        const data = [new ClipboardItem({[type]: blob})];

        try {
            await navigator.clipboard.write(data);
            popIn(copySuccess);
            clipboardModalTimeout = setTimeout(() => popOut(copySuccess), 1500);
        } catch (error) {
            popIn(copyError);
            clipboardModalTimeout = setTimeout(() => popOut(copyError), 1500);
        }
    }

    function outputAnimationEndAction(event) {
        const element = event.target;

        if (event.animationName === 'popout') {
            hide(element);
            element.classList.remove('animate-popout');
        }
    }

    /* ===== EVENT LISTENERS ===== */

    // Upload input events
    uploadBtn.addEventListener('keydown', event => event.key === "Enter" ? imageInput.click() : null);
    uploadBtn.addEventListener('drop', event => uploadBtnDropAction(event));
    uploadBtn.addEventListener('dragover', event => event.preventDefault());
    imageInput.addEventListener('change', event => imageInputChangeAction(event));

    // Size radio events
    sizeRadios.forEach(radio => {
        radio.addEventListener('click', event => sizeRadioClickAction(event));
    });
    for (const label of sizeRadioLabels) {
        label.addEventListener('keydown', event => sizeRadioLabelKeydownAction(event));
    }

    // Exposure input events
    exposure.addEventListener('input', event => {
        exposureValue.value = event.target.value;
    });
    exposureValue.addEventListener('change', event => {
        exposure.value = event.target.value;
    });

    // Form events
    form.addEventListener('submit', event => formSubmitAction(event));

    // Output events
    output.addEventListener('click', event => outputClickAction(event));
    copySuccess.addEventListener('animationend', event => outputAnimationEndAction(event));
    copyError.addEventListener('animationend', event => outputAnimationEndAction(event));
});