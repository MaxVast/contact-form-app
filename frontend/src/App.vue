<template>
  <main class="page">
    <ContactForm />
  </main>
</template>

<script setup>
import ContactForm from './components/ContactForm.vue'
</script>

<style>
* { box-sizing: border-box; }
body {
  margin: 0;
  font-family: system-ui, -apple-system, "Segoe UI", sans-serif;
  background: #f4f5f7;
}
.page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}
</style>
