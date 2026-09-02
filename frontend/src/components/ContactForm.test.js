import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import ContactForm from './ContactForm.vue'

function fillForm(wrapper) {
  return Promise.all([
    wrapper.find('input[type="text"]').setValue('Alice Dupont'),
    wrapper.find('input[type="email"]').setValue('alice@example.com'),
    wrapper.findAll('input[type="text"]')[1].setValue('Une question'),
    wrapper.find('textarea').setValue('Bonjour, ceci est un message de test.')
  ])
}

describe('ContactForm', () => {
  beforeEach(() => {
    global.fetch = vi.fn()
  })

  it('affiche les champs du formulaire', () => {
    const wrapper = mount(ContactForm)
    expect(wrapper.find('input[type="email"]').exists()).toBe(true)
    expect(wrapper.find('textarea').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').text()).toContain('Envoyer')
  })

  it('envoie le formulaire et affiche un message de succès', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: 1 })
    })

    const wrapper = mount(ContactForm)
    await fillForm(wrapper)
    await wrapper.find('form').trigger('submit.prevent')
    await vi.waitUntil(() => wrapper.find('.success').exists())

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/contact/',
      expect.objectContaining({ method: 'POST' })
    )
    expect(wrapper.find('.success').exists()).toBe(true)
    expect(wrapper.find('input[type="email"]').element.value).toBe('')
  })

  it('affiche un message d\'erreur si l\'API répond en erreur', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: false,
      json: async () => ({ error: 'email invalide' })
    })

    const wrapper = mount(ContactForm)
    await fillForm(wrapper)
    await wrapper.find('form').trigger('submit.prevent')
    await vi.waitUntil(() => wrapper.find('.error').exists())

    expect(wrapper.find('.error').text()).toContain('email invalide')
  })

  it('affiche un message d\'erreur générique si fetch échoue (réseau)', async () => {
    global.fetch.mockRejectedValueOnce(new Error('network down'))

    const wrapper = mount(ContactForm)
    await fillForm(wrapper)
    await wrapper.find('form').trigger('submit.prevent')
    await vi.waitUntil(() => wrapper.find('.error').exists())

    expect(wrapper.find('.error').text()).toContain('network down')
  })

  it('désactive le bouton pendant l\'envoi', async () => {
    let resolveFetch
    global.fetch.mockReturnValueOnce(
      new Promise((resolve) => { resolveFetch = resolve })
    )

    const wrapper = mount(ContactForm)
    await fillForm(wrapper)
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeDefined()

    resolveFetch({ ok: true, json: async () => ({}) })
    await vi.waitUntil(() => !wrapper.find('button[type="submit"]').attributes('disabled'))
  })
})
