<template>
  <div class="create">
    <div class="image-upload-section">
      <div
          class="image-upload-area"
          :class="{ 'error': uploadError, 'uploaded': uploadedImage }"
          @click="triggerFileUpload"
      >
        <input
            ref="fileInput"
            type="file"
            accept="image/*"
            @change="handleFileUpload"
            style="display: none"
        />

        <div v-if="!uploadedImage && !isUploading" class="upload-placeholder">
          <ImageSVG/>
          <div class="upload-text-wrapper">
            <div class="upload-text"><span>Click to upload</span> an image</div>
            <span class="upload-requirements">Max. 5MB, 325 x 240px</span>
          </div>
        </div>

        <div v-if="isUploading" class="upload-progress">
          <span>Uploading your image</span>
          <div class="progress-bar">
            <div class="progress-fill"></div>
          </div>
        </div>

        <div v-if="uploadedImage" class="uploaded-image">
          <img :src="uploadedImage" alt="Duel image"/>
          <div class="uploaded-image__button">
            <ButtonTag
                name="Upload Image"
                @click.stop="triggerFileUpload"
                variant="upload"
            >
              <template #icon>
                <LogOutSVG/>
              </template>
            </ButtonTag>
          </div>
        </div>
      </div>
      <div v-if="uploadError" class="error-message">{{ uploadError }}</div>
    </div>

    <div class="form-field">
      <label class="field-label">
        <span>Question</span>
        <span><span :class="{ 'error': questionError }">{{ formData.question.length }}</span>/200</span>
      </label>
      <Textarea
          v-model="formData.question"
          placeholder="Can Lina go one week without coffee?"
          maxlength="200"
          @input="validateQuestion"
          :error="questionError"
      />
      <div v-if="questionError" class="error-message">{{ questionError }}</div>
    </div>

    <div class="form-field">
      <label class="field-label">Deadline</label>
      <Input
          v-model="formData.event_date"
          type="datetime-local"
          @input="validateDate"
      >
        <template #icon>
          <CalendarSVG/>
        </template>
      </Input>
      <div v-if="deadlineError" class="error-message">{{ deadlineError }}</div>
    </div>

    <div class="form-field">
      <label class="field-label">Price</label>
      <Input
          v-model="formData.duel_price"
          placeholder="Price"
          type="number"
          inputmode="decimal"
          @input="validatePrice"
          :error="priceError"
      >
        <template #icon>
          <USDCSVG/>
        </template>
      </Input>
      <div v-if="priceError" class="error-message">{{ priceError }}</div>
    </div>

    <div class="form-field">
      <label class="field-label">Commission</label>
      <RangeInput
          v-model="formData.commission"
          min="0"
          max="20"
      />
    </div>

    <div class="form-field">
      <label class="field-label">Chosen one and confirm</label>
      <div class="answer-buttons">
        <Button
            :disabled="isLoading || !isFormValid"
            name="No"
            variant="red"
            @click="submitForm(0)"
        />
        <Button
            :disabled="isLoading || !isFormValid"
            name="Yes"
            variant="green"
            @click="submitForm(1)"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import {ref, computed, onMounted} from 'vue'
import ImageSVG from "../components/SVG/ImageSVG.vue";
import Button from "../components/UI/Button.vue";
import Input from "../components/UI/Input.vue";
import USDCSVG from "../components/SVG/USDCSVG.vue";
import Textarea from "../components/UI/Textarea.vue";
import RangeInput from "../components/UI/RangeInput.vue";
import LogOutSVG from "../components/SVG/LogOutSVG.vue";
import CalendarSVG from "../components/SVG/CalendarSVG.vue";
import {CreateDuel, SignCreateDuel, UploadFile} from "../api/index.js";
import {useWalletStore} from "../store/walletStore.js";
import {useRouter} from "vue-router";
import ButtonTag from "../components/UI/ButtonTag.vue";
import {useUserStore} from "../store/userStore.js";
import {useNotificationStore} from "../store/notificationStore.js";

const notificationStore = useNotificationStore();
const walletStore = useWalletStore();
const userStore = useUserStore();
const router = useRouter();

const isLoading = ref(false);

const formData = ref({
  bg_url: '',
  question: '',
  duel_price: 1,
  commission: 10,
  event_date: '',
})

const fileInput = ref(null)
const uploadedImage = ref('')
const uploadedFile = ref(null)
const isUploading = ref(false)
const uploadError = ref('')

const deadlineError = ref('')
const questionError = ref('')
const priceError = ref('')

const isFormValid = computed(() => {
  return formData.value.question.trim() &&
      formData.value.duel_price &&
      formData.value.event_date &&
      !questionError.value &&
      !priceError.value
})

const triggerFileUpload = () => {
  fileInput.value?.click()
}

const handleFileUpload = (event) => {
  const file = event.target.files[0]
  if (!file) return

  if (file.size > 5 * 1024 * 1024) {
    uploadError.value = 'File size exceeds 5MB limit'
    return
  }

  const img = new Image()
  img.onload = () => {
    // if (img.width !== 325 || img.height !== 240) {
    //   uploadError.value = 'Image dimensions must be exactly 325x240px'
    //   return
    // }

    const reader = new FileReader()
    reader.onload = (e) => {
      uploadedImage.value = e.target.result
      uploadedFile.value = file
      uploadError.value = ''
      isUploading.value = false
    }
    reader.readAsDataURL(file)
  }

  img.src = URL.createObjectURL(file)
  isUploading.value = true
  uploadError.value = ''
}

const validateQuestion = () => {
  if (formData.value.question.length > 200) {
    questionError.value = 'The question is too long'
  } else {
    questionError.value = ''
  }
}

const validatePrice = () => {
  const price = parseFloat(formData.value.duel_price)
  if (price > 1000000) {
    priceError.value = 'The price is too big'
  } else if (price > 0 && price < 1) {
    priceError.value = 'Min price is too low'
  } else if (price < 0) {
    priceError.value = 'Price cannot be negative'
  } else {
    priceError.value = ''
  }
}

const validateDate = () => {
  const deadline = new Date(formData.value.event_date);

  const now = new Date();

  if (deadline < now) {
    deadlineError.value = 'The deadline cannot be in the past'
  } else {
    deadlineError.value = ''
  }
}

const submitForm = async (answer) => {
  if (!isFormValid.value) return;
  isLoading.value = true;

  const id = String(Date.now() * Math.random());

  try {
    if (!userStore.userData) {
      await walletStore.connect();
      await submitForm(answer);
    } else {
      notificationStore.addNotification({
        type: 'loading',
        text: 'Please confirm transaction'
      }, 0, id);

      if (uploadedFile.value) {
        const data = new FormData();
        data.append('images', uploadedFile.value);
        const resp = await UploadFile(data);

        formData.value.bg_url = resp.data.image_urls[0];
      }

      const obj = {
        bg_url: formData.value.bg_url,
        question: formData.value.question,
        duel_price: parseFloat(formData.value.duel_price),
        commission: parseFloat(formData.value.commission),
        event_date: new Date(formData.value.event_date).toISOString(),
        answer: answer,
      }
      const resp = await SignCreateDuel(obj);

      const tx_hash = await walletStore.sendTx(resp.data?.tx);

      if (tx_hash) {
        setTimeout(async () => {
          await CreateDuel(obj, tx_hash);

          await walletStore.getBalance();

          await userStore.getResolveCount();

          await router.push({ name: 'home' });

          notificationStore.addNotification({
            type: 'success',
            text: 'You’ve successfully created the duel! <br> Now it’s time to wait for the results — good luck!'
          });
        })
      }
    }
  } catch (error) {
    notificationStore.addNotification({type: 'error', text: 'Somthing went wrong'});
  } finally {
    isLoading.value = false;
    notificationStore.removeNotification(id);
  }
}

onMounted(() => {
  const now = new Date()
  now.setDate(now.getDate() + 1);

  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')

  formData.value.event_date = `${year}-${month}-${day}T${hours}:${minutes}`
})
</script>

<style scoped lang="scss">
.create {
}

.form-field {
  margin-bottom: 20px;
  &:last-child {
    margin-bottom: 0;
  }
}

.field-label {
  color: #C9CCD8;
  font-size: 12px;
  font-weight: 300;
  line-height: 18px;
  letter-spacing: 0.2px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;

  .error {
    color: #E44E2D;
  }
}

.error-message {
  color: #E44E2D;
  font-size: 12px;
  font-weight: 300;
  line-height: 18px;
  letter-spacing: 0.2px;
  margin-top: 4px;
}

.image-upload-section {
  margin-bottom: 16px;
}

.image-upload-area {
  width: 100%;
  height: 215px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: 0.3s;
  border-radius: 12px;
  background: #1C1C1C;
  border: 1px solid transparent;
  padding: 22px;
  overflow: hidden;

  &.error {
    border: 1px solid rgba(238, 68, 68, 0.89);
  }

  &.uploaded {
    padding: 0;
    border: none;
  }
}

.upload-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.upload-text {
  color: #F9F8F8;
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
  letter-spacing: 0.2px;

  span {
    color: #D0F267;
  }

  &-wrapper {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
}

.upload-requirements {
  color: #C9CCD8;
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
  letter-spacing: 0.2px;
}

.upload-progress {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: #F9F8F8;
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
  letter-spacing: 0.2px;
  width: 100%;
}

.progress-bar {
  width: 100%;
  height: 6px;
  overflow: hidden;
  border-radius: 12px;
  border: 1px solid #292929;
  background: #1C1C1C;
}

.progress-fill {
  height: 100%;
  border-radius: 12px;
  background: linear-gradient(290deg, #F1B772 -0.67%, #4CB673 54.38%, #F1E367 100%);
  animation: progress 2s ease-in-out infinite;
}

@keyframes progress {
  0% {
    width: 0%;
  }
  50% {
    width: 70%;
  }
  100% {
    width: 100%;
  }
}

.uploaded-image {
  position: relative;
  width: 100%;
  height: 100%;

  &::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    width: 100%;
    height: 85px;
    background: linear-gradient(180deg, rgba(0, 0, 0, 0.00) 14.86%, #0A1214 100%);
    z-index: 11;
  }

  img {
    border-radius: 12px;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  &__button {
    position: absolute;
    bottom: 16px;
    left: 16px;
    z-index: 100;
  }
}

.answer-buttons {
  display: flex;
  gap: 12px;
}
</style>
