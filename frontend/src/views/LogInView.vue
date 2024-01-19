<template>
    <div class="logInContainer">
            <h2>Enter your recovery phrase</h2>
            <form @submit.prevent="submitForm">
            <div v-for="(i,k) in recoveryPhrase" v-bind:key="k" class="secretPhrase">
                <p>{{ k+1 }}</p>
                <input type="password" v-model="recoveryPhrase[k]">
            </div>
            <button type="submit">Log In</button>
        </form>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const submitForm = async () => {
    try {
        let request = {
            recoveryPhrase: ""
        }
        for (let i = 0; i < recoveryPhrase.value.length; i++) {
            request.recoveryPhrase += recoveryPhrase.value[i] + " "
        }
        const response = await fetch('http://localhost:3000/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(request)
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
    } catch (error) {
        console.error(error)
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
h2{
    margin-bottom: 25px;
}
.logInContainer{
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    height: 100vh;
}
form{
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 20px;
}

form>button{
    grid-column-start:2;
    padding: 5px;
    border: none;
    font-size: 16px;
    background-color: var(--accent-color);
    border-radius: 5px;
    transition: 0.3s;
}
form>button:hover{
    transform: scale(1.1);
    transition: 0.3s;
}
@media (max-width: 768px) {
    form{
        grid-template-columns: 1fr 1fr;
    }
    form>button{
        grid-column-start:1;
        grid-column-end:3;
    }
}
.secretPhrase{
    position: relative;
}
.secretPhrase>p{
    position: absolute;
    left:-16px;
    text-align:end;
    width: 16px;
    font-size: 14px;
}
.secretPhrase>input{
    width: 150px;
    padding: 5px;
    font-size: 16px;
    border: 2px solid var(--black-color);
    border-radius: 5px;
}
.secretPhrase>input:focus{
    outline: none;
}
</style>

