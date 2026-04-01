<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import type { DancerProfile } from './SwipeCard.svelte';

	interface Props {
		open?: boolean;
		myPhoto?: string;
		theirProfile?: DancerProfile | null;
		onchat?: () => void;
		onclose?: () => void;
	}

	let { open = false, myPhoto = '', theirProfile = null, onchat, onclose }: Props = $props();
</script>

{#if open}
	<div
		class="fixed inset-0 z-[100] flex flex-col items-center overflow-hidden"
		style="background: linear-gradient(75.07deg, rgb(137,132,218) 13.4%, rgb(196,182,241) 86.6%);"
		transition:fade={{ duration: 250 }}
	>
		<!-- Close button -->
		<button
			class="absolute top-[54px] left-4 flex h-[38px] w-[38px] items-center justify-center"
			onclick={onclose}
			aria-label="Close"
		>
			<i class="fi fi-rr-cross text-white" style="font-size: 24px; line-height: 1;"></i>
		</button>

		<!-- Content -->
		<div
			class="flex flex-col items-center"
			style="margin-top: max(env(safe-area-inset-top), 54px); gap: 40px; padding: 0 16px;"
			transition:scale={{ start: 0.9, duration: 350, delay: 100 }}
		>
			<!-- Heading -->
			<h1 class="text-center text-[40px] font-black text-white leading-tight">You matched!</h1>

			<!-- Photos + logo -->
			<div class="relative flex items-center justify-center" style="height: 200px; width: 343px;">
				<!-- Left photo (tilted -9.46°) -->
				<div
					class="absolute overflow-hidden rounded-[20px]"
					style="width: 160px; height: 200px; transform: rotate(-9.46deg); left: 0;"
				>
					<img
						src={myPhoto || '/placeholder-avatar.jpg'}
						alt="You"
						class="h-full w-full object-cover"
					/>
				</div>
				<!-- Right photo (tilted +9.04°) -->
				<div
					class="absolute overflow-hidden rounded-[20px]"
					style="width: 160px; height: 200px; transform: rotate(9.04deg); right: 0;"
				>
					<img
						src={theirProfile?.photoUrl || '/placeholder-avatar.jpg'}
						alt={theirProfile?.name}
						class="h-full w-full object-cover"
					/>
				</div>
				<!-- MatchUp logo overlay -->
				<div
					class="absolute flex h-[90px] w-[101px] items-center justify-center rounded-[20px]"
					style="background: linear-gradient(180deg, #c8c8c8 0%, #888 100%); bottom: -10px; left: 50%; transform: translateX(-50%); z-index: 10;"
				>
					<span class="text-[10px] font-semibold text-white">MatchUp</span>
				</div>
			</div>

			<!-- Subtitle -->
			<div class="flex flex-col items-center gap-1 text-center">
				<p class="text-[20px] font-normal text-white">
					You and {theirProfile?.name} like each other.
				</p>
				<p class="text-[20px] font-normal text-white" style="letter-spacing: 1px;">
					Why not make the first move?
				</p>
			</div>

			<!-- Chat now button -->
			<button
				class="flex items-center justify-center rounded-[50px] bg-white px-4 mix-blend-plus-lighter"
				style="height: 40px; width: 163px;"
				onclick={onchat}
			>
				<span class="text-[14px] font-semibold" style="color: #171717;">Chat now</span>
			</button>
		</div>
	</div>
{/if}
