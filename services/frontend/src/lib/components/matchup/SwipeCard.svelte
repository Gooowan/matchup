<script lang="ts">
	import { spring } from 'svelte/motion';
	import { t } from '$lib/locale';

	export const STOCK_AVATAR = 'https://images.unsplash.com/photo-1545959570-a94084071b5d?w=800&auto=format';

	export interface DancerProfile {
		id: string;
		name: string;
		age: number;
		photoUrl: string;
		tags: string[];       // e.g. ["Ballroom", "Pro", "1.75 cm"]
		location: string;
		school?: string;
		goals?: string;
	}

	interface Props {
		profile: DancerProfile;
		onlike?: (id: string) => void;
		onpass?: (id: string) => void;
		onviewprofile?: (id: string) => void;
		onmenu?: (id: string) => void;
		zIndex?: number;
		isTop?: boolean;
		fullScreen?: boolean;
	}

	let {
		profile,
		onlike,
		onpass,
		onviewprofile,
		onmenu,
		zIndex = 0,
		isTop = false,
		fullScreen = false
	}: Props = $props();

	// Spring-animated card position
	const pos = spring({ x: 0, y: 0, rot: 0 }, { stiffness: 0.15, damping: 0.85 });

	let dragging = $state(false);
	let startX = 0;
	let startY = 0;
	let cardEl: HTMLElement;

	const SWIPE_THRESHOLD = 120;

	// Overlay opacity tied to drag distance
	let swipeDir = $derived(
		$pos.x > 40 ? 'right' : $pos.x < -40 ? 'left' : null
	);
	let overlayOpacity = $derived(Math.min(Math.abs($pos.x) / SWIPE_THRESHOLD, 1));

	function onPointerDown(e: PointerEvent) {
		if (!isTop) return;
		dragging = true;
		startX = e.clientX - $pos.x;
		startY = e.clientY - $pos.y;
		cardEl.setPointerCapture(e.pointerId);
	}

	function onPointerMove(e: PointerEvent) {
		if (!dragging) return;
		const dx = e.clientX - startX;
		const dy = e.clientY - startY;
		pos.set({ x: dx, y: dy, rot: dx / 20 }, { hard: false });
	}

	function onPointerUp(_e: PointerEvent) {
		if (!dragging) return;
		dragging = false;

		if ($pos.x > SWIPE_THRESHOLD) {
			commitSwipe('right');
		} else if ($pos.x < -SWIPE_THRESHOLD) {
			commitSwipe('left');
		} else {
			// Snap back
			pos.set({ x: 0, y: 0, rot: 0 });
		}
	}

	function commitSwipe(dir: 'right' | 'left') {
		const targetX = dir === 'right' ? window.innerWidth + 200 : -(window.innerWidth + 200);
		pos.set({ x: targetX, y: $pos.y, rot: dir === 'right' ? 15 : -15 }, { hard: false });
		setTimeout(() => {
			if (dir === 'right') onlike?.(profile.id);
			else onpass?.(profile.id);
		}, 300);
	}

	// Button-triggered swipes
	export function swipeRight() { commitSwipe('right'); }
	export function swipeLeft() { commitSwipe('left'); }
</script>

<!-- Card wrapper with spring transform -->
<div
	bind:this={cardEl}
	class="absolute"
	style="
		{fullScreen ? 'inset: 0;' : 'width: 343px; height: 581px; left: 16px; top: 114px;'}
		z-index: {zIndex};
		transform: translate({$pos.x}px, {$pos.y}px) rotate({$pos.rot}deg);
		touch-action: none;
		cursor: {isTop ? 'grab' : 'default'};
		will-change: transform;
	"
	onpointerdown={onPointerDown}
	onpointermove={onPointerMove}
	onpointerup={onPointerUp}
	onpointercancel={onPointerUp}
	role="button"
	tabindex="0"
	aria-label="Profile card for {profile.name}"
>
	<!-- Card itself -->
	<div class="relative h-full w-full overflow-hidden" style="border-radius: {fullScreen ? '0' : '20px'}">
		<!-- Photo -->
		<img
			src={profile.photoUrl || STOCK_AVATAR}
			alt={profile.name}
			class="absolute inset-0 h-auto w-full object-cover"
			style="min-height: 100%; object-position: center top;"
			draggable="false"
			onerror={(e) => { (e.currentTarget as HTMLImageElement).src = STOCK_AVATAR; }}
		/>

		<!-- Top gradient -->
		<div
			class="absolute top-0 right-0 left-0 h-[124px]"
			style="background: linear-gradient(180deg, rgba(0,0,0,0.3) 0%, rgba(0,0,0,0) 100%);"
		></div>

		<!-- Bottom gradient -->
		<div
			class="absolute right-0 bottom-0 left-0 h-[295px]"
			style="background: linear-gradient(0deg, rgba(0,0,0,1) 0%, rgba(0,0,0,0) 100%);"
		></div>

		<!-- LIKE overlay (swipe right) -->
		{#if swipeDir === 'right'}
			<div
				class="absolute inset-0"
				style="
					background: radial-gradient(ellipse 750px 812px at -375px 50%, rgba(142,161,223,0.5) 0%, transparent 70%);
					opacity: {overlayOpacity};
				"
			></div>
			<div
				class="absolute text-white"
				style="left: 40px; top: 50%; transform: translateY(-50%); opacity: {overlayOpacity};"
			>
				<i class="fi fi-rr-heart" style="font-size: 60px; line-height: 1;"></i>
			</div>
		{/if}

		<!-- PASS overlay (swipe left) -->
		{#if swipeDir === 'left'}
			<div
				class="absolute inset-0"
				style="
					background: radial-gradient(ellipse 750px 812px at 750px 50%, rgba(251,194,235,0.5) 0%, transparent 70%);
					opacity: {overlayOpacity};
				"
			></div>
			<div
				class="absolute text-white"
				style="right: 40px; top: 50%; transform: translateY(-50%); opacity: {overlayOpacity};"
			>
				<i class="fi fi-rr-cross" style="font-size: 60px; line-height: 1;"></i>
			</div>
		{/if}

		<!-- 3-dot menu button -->
		<button
			class="absolute flex items-center justify-center rounded-full"
			style="width: 38px; height: 38px; right: 8px; top: 8px;"
			onclick={(e) => { e.stopPropagation(); onmenu?.(profile.id); }}
			aria-label={$t('feed.action_more_options')}
		>
			<i class="fi fi-rr-menu-dots-vertical rotate-90 text-white" style="font-size: 15px; line-height: 1;"></i>
		</button>

		<!-- Profile info overlay -->
		<div
			class="absolute right-4 left-4 flex flex-col"
			style="bottom: calc(max(env(safe-area-inset-bottom), 8px) + 80px); gap: 24px;"
		>
			<!-- Name -->
			<div class="flex items-baseline gap-0">
				<span class="text-[20px] font-black leading-tight text-white tracking-tight"
					>{profile.name}</span>
				<span class="text-[20px] font-black leading-tight text-white" style="letter-spacing: -1px;"
					>, {profile.age}</span>
			</div>

			<!-- Tag pills -->
			<div class="flex flex-wrap gap-2">
				{#each profile.tags as tag}
					<span class="card-pill">{tag}</span>
				{/each}
			</div>

			<!-- Info rows -->
			<div class="flex flex-col gap-2">
				{#if profile.location}
					<div class="flex items-center gap-1.5">
						<i class="fi fi-rr-marker text-white" style="font-size: 15px; line-height: 1; flex-shrink: 0;"></i>
						<span class="text-[12px] font-medium text-white">{profile.location}</span>
					</div>
				{/if}
				{#if profile.school}
					<div class="flex items-center gap-1.5">
						<i class="fi fi-rr-bank text-white" style="font-size: 15px; line-height: 1; flex-shrink: 0;"></i>
						<span class="text-[12px] font-medium text-white">{profile.school}</span>
					</div>
				{/if}
				{#if profile.goals}
					<div class="flex items-start gap-1.5">
						<i class="fi fi-rr-star text-white" style="font-size: 15px; line-height: 1; flex-shrink: 0; margin-top: 1px;"></i>
						<span class="text-[12px] font-medium text-white leading-snug">{profile.goals}</span>
					</div>
				{/if}
			</div>

			<!-- Action buttons -->
			<div class="flex items-center gap-2">
				<!-- Pass -->
				<button
					class="flex flex-1 items-center justify-center gap-1.5 rounded-[50px] mix-blend-plus-lighter"
					style="height: 38px; background: linear-gradient(43.36deg, rgb(142,161,223) 2%, rgba(142,161,223,0) 41.6%), white;"
					onclick={() => swipeLeft()}
				>
					<i class="fi fi-rr-cross" style="font-size: 15px; line-height: 1; color: #171717;"></i>
					<span class="text-[12px] font-semibold" style="color: #171717;">{$t('swipe.pass')}</span>
				</button>

				<!-- View Profile -->
				<button
					class="flex items-center justify-center rounded-[50px] bg-white px-4 mix-blend-plus-lighter"
					style="height: 38px;"
					onclick={() => onviewprofile?.(profile.id)}
				>
					<span class="text-[12px] font-semibold" style="color: #171717;">{$t('swipe.view_profile')}</span>
				</button>

				<!-- Like -->
				<button
					class="flex flex-1 items-center justify-center gap-1.5 rounded-[50px] mix-blend-plus-lighter"
					style="height: 38px; background: linear-gradient(-41.99deg, rgb(251,194,235) 0.2%, rgba(251,194,235,0) 50.5%), white;"
					onclick={() => swipeRight()}
				>
					<i class="fi fi-rr-heart" style="font-size: 15px; line-height: 1; color: #171717;"></i>
					<span class="text-[12px] font-semibold" style="color: #171717;">{$t('swipe.like')}</span>
				</button>
			</div>
		</div>
	</div>
</div>
