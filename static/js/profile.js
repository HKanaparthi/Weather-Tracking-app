document.addEventListener('DOMContentLoaded', function() {
    // Initialize photo selection functionality
    initPhotoSelection();

    // Initialize avatar color selection
    initAvatarColorSelection();

    // Get form elements
    const profileForm = document.querySelector('form');
    const inputs = document.querySelectorAll('.form-input');
    const passwordFields = document.querySelectorAll('input[type="password"]');

    // Add focus effects for all form inputs
    inputs.forEach(input => {
        // Add focus and blur events for visual feedback
        input.addEventListener('focus', function() {
            this.classList.add('input-active');
        });

        input.addEventListener('blur', function() {
            this.classList.remove('input-active');

            // Validate required fields
            if (this.hasAttribute('required') && !this.value.trim()) {
                this.classList.add('input-error');
                showErrorMessage(this, 'This field is required');
            } else {
                this.classList.remove('input-error');
                removeErrorMessage(this);
            }
        });
    });

    // Password validation
    if (passwordFields.length > 0) {
        const currentPassword = document.getElementById('current_password');
        const newPassword = document.getElementById('new_password');

        // Handle password change validation
        passwordFields.forEach(field => {
            field.addEventListener('input', function() {
                // If one password field has value, make sure both have values
                const otherField = this === currentPassword ? newPassword : currentPassword;

                if (this.value && !otherField.value) {
                    showErrorMessage(otherField, 'Both password fields are required to change password');
                } else {
                    removeErrorMessage(otherField);
                }

                // Password strength indicator (only for new password)
                if (this === newPassword && this.value) {
                    showPasswordStrength(this);
                }
            });
        });
    }

    // Form submission
    if (profileForm) {
        profileForm.addEventListener('submit', function(e) {
            let hasErrors = false;

            // Check all required fields
            const requiredInputs = this.querySelectorAll('[required]');
            requiredInputs.forEach(input => {
                if (!input.value.trim()) {
                    input.classList.add('input-error');
                    showErrorMessage(input, 'This field is required');
                    hasErrors = true;
                }
            });

            // Check password fields consistency
            if (passwordFields.length > 0) {
                const currentPassword = document.getElementById('current_password');
                const newPassword = document.getElementById('new_password');

                if ((currentPassword.value && !newPassword.value) || (!currentPassword.value && newPassword.value)) {
                    if (!currentPassword.value) {
                        currentPassword.classList.add('input-error');
                        showErrorMessage(currentPassword, 'Both password fields are required to change password');
                    }
                    if (!newPassword.value) {
                        newPassword.classList.add('input-error');
                        showErrorMessage(newPassword, 'Both password fields are required to change password');
                    }
                    hasErrors = true;
                }
            }

            if (hasErrors) {
                e.preventDefault();

                // Smooth scroll to the first error
                const firstError = document.querySelector('.input-error');
                if (firstError) {
                    firstError.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    firstError.focus();
                }
            }
        });
    }

    // Check for success notification
    checkForSuccessMessage();

    // Helper functions
    function showErrorMessage(element, message) {
        // Remove existing error message if any
        removeErrorMessage(element);

        // Create new error message
        const errorMsg = document.createElement('div');
        errorMsg.className = 'form-error-message';
        errorMsg.textContent = message;

        // Insert after the input element
        element.parentNode.insertBefore(errorMsg, element.nextSibling);
    }

    function removeErrorMessage(element) {
        const existingError = element.parentNode.querySelector('.form-error-message');
        if (existingError) {
            existingError.remove();
        }
    }

    function showPasswordStrength(passwordField) {
        // Remove existing strength indicator
        const existingIndicator = passwordField.parentNode.querySelector('.password-strength');
        if (existingIndicator) {
            existingIndicator.remove();
        }

        // Calculate strength
        const strength = calculatePasswordStrength(passwordField.value);

        // Create strength indicator
        const strengthIndicator = document.createElement('div');
        strengthIndicator.className = `password-strength ${strength.level}`;

        strengthIndicator.innerHTML = `
            <div class="strength-bar">
                <div class="strength-fill" style="width: ${strength.percentage}%"></div>
            </div>
            <div class="strength-text">${strength.message}</div>
        `;

        // Insert after password field
        passwordField.parentNode.insertBefore(strengthIndicator, passwordField.nextSibling);
    }

    function calculatePasswordStrength(password) {
        let score = 0;

        // Award points for length
        if (password.length >= 8) score += 25;
        if (password.length >= 12) score += 15;

        // Award points for complexity
        if (/[a-z]/.test(password)) score += 10; // lowercase
        if (/[A-Z]/.test(password)) score += 15; // uppercase
        if (/[0-9]/.test(password)) score += 15; // numbers
        if (/[^a-zA-Z0-9]/.test(password)) score += 20; // special characters

        // Cap at 100
        score = Math.min(score, 100);

        // Determine level and message
        let level, message;

        if (score < 30) {
            level = 'weak';
            message = 'Weak password';
        } else if (score < 60) {
            level = 'medium';
            message = 'Medium strength';
        } else if (score < 80) {
            level = 'strong';
            message = 'Strong password';
        } else {
            level = 'very-strong';
            message = 'Very strong password';
        }

        return {
            percentage: score,
            level: level,
            message: message
        };
    }

    // Photo selection functionality
    function initPhotoSelection() {
        const photoOptions = document.querySelectorAll('.photo-option');
        const currentPhoto = document.getElementById('current-photo');
        const profilePhotoInput = document.getElementById('profile_photo');

        if (!photoOptions.length || !currentPhoto || !profilePhotoInput) {
            console.warn('Photo selection elements not found');
            return;
        }

        photoOptions.forEach(option => {
            option.addEventListener('click', function() {
                // Remove selected class from all options
                photoOptions.forEach(opt => opt.classList.remove('selected'));

                // Add selected class to clicked option
                this.classList.add('selected');

                // Get photo filename from data attribute
                const photoFile = this.getAttribute('data-photo');

                // Update hidden input for form submission
                profilePhotoInput.value = photoFile;

                // Update the current photo display
                currentPhoto.src = `/static/profile_photos/${photoFile}`;

                // Add subtle animation to indicate change
                currentPhoto.classList.add('photo-changed');
                setTimeout(() => {
                    currentPhoto.classList.remove('photo-changed');
                }, 700);
            });
        });
    }

    // Avatar color selection
    function initAvatarColorSelection() {
        const avatarOptions = document.querySelectorAll('.avatar-option');
        const currentAvatar = document.getElementById('current-avatar');
        const avatarColorInput = document.getElementById('avatar_color');

        if (!avatarOptions.length || !currentAvatar || !avatarColorInput) {
            console.warn('Avatar selection elements not found');
            return;
        }

        avatarOptions.forEach(option => {
            option.addEventListener('click', function() {
                // Remove selected class from all options
                avatarOptions.forEach(opt => opt.classList.remove('selected'));

                // Add selected class to clicked option
                this.classList.add('selected');

                // Get color from data attribute
                const color = this.getAttribute('data-color');

                // Update hidden input for form submission
                avatarColorInput.value = color;

                // Remove all color classes from current avatar
                currentAvatar.classList.remove('avatar-red', 'avatar-blue', 'avatar-green',
                    'avatar-purple', 'avatar-orange', 'avatar-pink');

                // Add selected color class
                currentAvatar.classList.add(`avatar-${color}`);

                // Add subtle animation to indicate change
                currentAvatar.classList.add('avatar-changed');
                setTimeout(() => {
                    currentAvatar.classList.remove('avatar-changed');
                }, 700);
            });
        });
    }

    // Success notification handling
    function checkForSuccessMessage() {
        // Check if we need to show success notification on page load
        // This happens when we're returning from a successful form submission
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('success')) {
            showSuccessNotification();
        }

        // Also check if server returned a success message
        const serverSuccessMessage = document.querySelector('.success-message');
        if (serverSuccessMessage) {
            // Hide server message after 3 seconds
            setTimeout(function() {
                serverSuccessMessage.style.display = 'none';
            }, 3000);
        }
    }

    function showSuccessNotification() {
        const successNotification = document.getElementById('success-notification');
        if (!successNotification) return;

        successNotification.classList.add('show');

        // Hide notification after 3 seconds
        setTimeout(function() {
            successNotification.classList.remove('show');
        }, 3000);
    }

    // Add custom styles
    addCustomStyles();

    function addCustomStyles() {
        const style = document.createElement('style');
        style.textContent = `
            .input-active {
                border-color: rgba(255, 255, 255, 0.8) !important;
                box-shadow: 0 0 0 2px rgba(93, 156, 236, 0.3) !important;
            }
            
            .input-error {
                border-color: rgba(231, 76, 60, 0.8) !important;
                box-shadow: 0 0 0 2px rgba(231, 76, 60, 0.3) !important;
            }
            
            .form-error-message {
                color: rgba(255, 160, 160, 1);
                font-size: 0.85rem;
                margin-top: 5px;
                animation: fadeIn 0.3s ease;
            }
            
            @keyframes fadeIn {
                from { opacity: 0; transform: translateY(-5px); }
                to { opacity: 1; transform: translateY(0); }
            }
            
            .password-strength {
                margin-top: 10px;
                font-size: 0.85rem;
            }
            
            .strength-bar {
                height: 4px;
                background-color: rgba(255, 255, 255, 0.2);
                border-radius: 2px;
                margin-bottom: 5px;
                overflow: hidden;
            }
            
            .strength-fill {
                height: 100%;
                transition: width 0.3s ease;
            }
            
            .strength-text {
                font-size: 0.75rem;
                color: rgba(255, 255, 255, 0.8);
            }
            
            .weak .strength-fill {
                background-color: rgba(231, 76, 60, 0.8);
            }
            
            .medium .strength-fill {
                background-color: rgba(243, 156, 18, 0.8);
            }
            
            .strong .strength-fill {
                background-color: rgba(93, 156, 236, 0.8);
            }
            
            .very-strong .strength-fill {
                background-color: rgba(46, 204, 113, 0.8);
            }
            
            /* Profile photo styles */
            .photo-option {
                cursor: pointer;
                transition: all 0.2s ease;
                border: 2px solid transparent;
            }
            
            .photo-option:hover {
                transform: scale(1.05);
                box-shadow: 0 0 10px rgba(255, 255, 255, 0.3);
            }
            
            .photo-option.selected {
                border-color: #4e54c8;
                box-shadow: 0 0 10px rgba(78, 84, 200, 0.5);
            }
            
            @keyframes photoChanged {
                0% { transform: scale(1); }
                50% { transform: scale(1.05); }
                100% { transform: scale(1); }
            }
            
            .photo-changed {
                animation: photoChanged 0.7s ease;
            }
            
            @keyframes avatarChanged {
                0% { transform: rotate(0deg); }
                25% { transform: rotate(-5deg); }
                75% { transform: rotate(5deg); }
                100% { transform: rotate(0deg); }
            }
            
            .avatar-changed {
                animation: avatarChanged 0.7s ease;
            }
        `;
        document.head.appendChild(style);
    }
});