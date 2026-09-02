<template>
  <form class="card" @submit.prevent="handleSubmit" novalidate>
    <h1>Nous contacter</h1>
    <p class="subtitle">Une question, un projet ? Écrivez-nous.</p>

    <label>
      Nom
      <input v-model.trim="form.name" type="text" maxlength="100" required />
    </label>

    <label>
      Email
      <input v-model.trim="form.email" type="email" maxlength="150" required />
    </label>

    <label>
      Sujet
      <input v-model.trim="form.subject" type="text" maxlength="150" required />
    </label>

    <label>
      Message
      <textarea v-model.trim="form.message" rows="5" maxlength="5000" required />
    </label>

    <button type="submit" :disabled="status === 'loading'">
      {{ status === 'loading' ? 'Envoi…' : 'Envoyer' }}
    </button>

    <p v-if="status === 'success'" class="msg success">
      Message envoyé, merci ! Nous revenons vers vous rapidement.
    </p>
    <p v-if="status === 'error'" class="msg error">{{ errorMessage }}</p>
  </form>
</template>

<script setup>
import { reactive, ref } from 'vue'

const form = reactive({ name: '', email: '', subject: '', message: '' })
const status = ref('idle') // idle | loading | success | error
const errorMessage = ref('')

async function handleSubmit() {
  status.value = 'loading'
  errorMessage.value = ''
  try {
    const res = await fetch('/api/contact/', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(form)
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || "L'envoi a échoué, réessayez.")
    }

    status.value = 'success'
    form.name = form.email = form.subject = form.message = ''
  } catch (err) {
    status.value = 'error'
    errorMessage.value = err.message
  }
}
</script>

<style scoped>
.card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  padding: 32px;
  width: 100%;
  max-width: 480px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
h1 { margin: 0; font-size: 1.5rem; }
.subtitle { margin: 0 0 8px; color: #666; }
label { display: flex; flex-direction: column; gap: 6px; font-size: 0.9rem; color: #333; }
input, textarea {
  padding: 10px 12px;
  border: 1px solid #d7d9dd;
  border-radius: 8px;
  font: inherit;
  resize: vertical;
}
input:focus, textarea:focus { outline: 2px solid #4f46e5; border-color: transparent; }
button {
  margin-top: 8px;
  padding: 12px;
  border: none;
  border-radius: 8px;
  background: #4f46e5;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
}
button:disabled { opacity: 0.6; cursor: not-allowed; }
.msg { margin: 0; font-size: 0.9rem; }
.success { color: #15803d; }
.error { color: #b91c1c; }
</style>
