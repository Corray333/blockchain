<template>
  <div class="home">
    <h1>Blockchain</h1>
    <TransactionCard :block="block"/>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import TransactionCard from '../components/TransactionCard.vue'
let block = ref()
const check = async () => { 
    try { 
        const resp = await axios.get('http://localhost:3000/blocks/none') 
        block.value = resp.data
    } catch (error) { 
        alert("Ошибка", error)
    } 
}
onMounted(() => {
  check()
})

</script>
