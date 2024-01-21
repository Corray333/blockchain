<template>
    <navbar>
        <router-link to="/">Home</router-link>
        <router-link to="/login">LogIn</router-link>
    </navbar>
</template>

<script setup>
</script>

<style>

navbar {
    display: flex;
    background-color: var(--accent-color);
    width: 100%;
    color:var(--black-color);
    padding: 15px 150px;
    gap: 15px;
}

navbar>a{
    transition: .3s;
}

navbar>a:hover{
    transform: scale(1.1);
    transition: .3s;
}

</style>

