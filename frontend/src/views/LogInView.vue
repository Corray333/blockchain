<template>
    <div class="logInContainer">
        <h2>Enter your recovery phrase</h2>
        <form @submit.prevent="submitForm">
            <div v-for="(i, k) in recoveryPhrase" v-bind:key="k" class="secretPhrase">
                <p>{{ k + 1 }}</p>
                <input type="password" v-model="recoveryPhrase[k]">
            </div>
            <button type="submit">Log In</button>
        </form>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const submitForm = async () => { 
    try { 
        let request = { 
            recoveryPhrase: "" 
        } 
        for (let i = 0; i < recoveryPhrase.value.length; i++) { 
            request.recoveryPhrase += recoveryPhrase.value[i] + " " 
        } 
        const resp = await axios.post('http://localhost:3000/login', request) 
        if (resp.status == 200){
            localStorage.setItem('recPhrase', request.recoveryPhrase)
        }
    } catch (error) { 
        alert("Введен неверный секретный ключ")
    } 
}
    

const recoveryPhrase = ref([])
onMounted(() => {
    for (let i = 0; i < 12; i++) {
        recoveryPhrase.value.push('')
    }
})

</script>

<style>
h2 {
    margin-bottom: 25px;
}

.logInContainer {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: 100px;
}

form {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 20px;
}

form>button {
    grid-column-start: 2;
    padding: 5px;
    border: none;
    font-size: 16px;
    background-color: var(--accent-color);
    border-radius: 5px;
    transition: 0.3s;
}

form>button:hover {
    transform: scale(1.1);
    transition: 0.3s;
}

@media (max-width: 768px) {
    form {
        grid-template-columns: 1fr 1fr;
    }

    form>button {
        grid-column-start: 1;
        grid-column-end: 3;
    }
}

.secretPhrase {
    position: relative;
}

.secretPhrase>p {
    position: absolute;
    left: -17px;
    text-align: end;
    width: 16px;
    font-size: 14px;
}

.secretPhrase>input {
    width: 150px;
    padding: 5px;
    font-size: 16px;
    border: 2px solid var(--black-color);
    border-radius: 5px;
}

.secretPhrase>input:focus {
    outline: none;
}</style>

