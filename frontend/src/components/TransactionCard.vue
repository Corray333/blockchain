<template>
    <div class="block">
            <h3>{{ props.block.Timestamp }}</h3>
    </div>
</template>

<script setup>
import { defineProps } from 'vue'
const props = defineProps(['block'])

</script>

<style>

</style>