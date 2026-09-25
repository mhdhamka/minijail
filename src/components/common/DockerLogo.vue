<script setup lang="ts">
interface Props {
  className?: string;
  size?: number;
  monochrome?: boolean;
}

withDefaults(defineProps<Props>(), {
  className: 'w-7 h-7',
  size: 28,
  monochrome: false
});
</script>

<template>
  <svg 
    :class="className" 
    :width="size" 
    :height="size" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- Docker Container Boxes on Whale's Back -->
    <!-- Row 1 (top) -->
    <rect x="9.5" y="4" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />
    <rect x="12" y="4" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />

    <!-- Row 2 (middle) -->
    <rect x="7" y="6.5" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />
    <rect x="9.5" y="6.5" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />
    <rect x="12" y="6.5" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />
    <rect x="14.5" y="6.5" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />

    <!-- Row 3 (bottom containers) -->
    <rect x="4.5" y="9" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />
    <rect x="7" y="9" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />
    <rect x="9.5" y="9" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />
    <rect x="12" y="9" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />
    <rect x="14.5" y="9" width="2" height="2" rx="0.3" :fill="monochrome ? 'currentColor' : '#0db7ed'" />

    <!-- Whale Body -->
    <path 
      d="M21.8 11.2C21.5 11.1 20.8 11 20 11.4C19.5 10.6 18.7 10.1 17.8 10.1H17V12H19C19.4 12 19.8 12.1 20.1 12.3C19.5 13.5 18 14.5 16 14.8C15 15 11.5 15 9.5 14.2C7 13.2 5 12 2.5 12C2.1 12 1.6 12.1 1.2 12.2C1.1 12.7 1 13.2 1 13.8C1 17.2 4.1 19.5 9 19.5C15.8 19.5 21 16.5 22.5 12.2C22.6 11.8 22.2 11.4 21.8 11.2Z" 
      :fill="monochrome ? 'currentColor' : '#1D63ED'" 
    />

    <!-- Whale Eye -->
    <circle cx="18.5" cy="13" r="0.6" fill="#FFFFFF" />

    <!-- Water spout ripple -->
    <path 
      d="M22 9.5C22.5 8.5 23.5 8 24 8" 
      :stroke="monochrome ? 'currentColor' : '#0db7ed'" 
      stroke-width="0.8" 
      stroke-linecap="round" 
    />
  </svg>
</template>
