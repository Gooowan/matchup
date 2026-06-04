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
	let committed = false;
	let didDrag = false;
	let startX = 0;
	let startY = 0;
	let cardEl: HTMLElement;

	const SWIPE_THRESHOLD = 120;
	const TAP_THRESHOLD = 8;

	// Overlay opacity tied to drag distance
	let swipeDir = $derived(
		$pos.x > 40 ? 'right' : $pos.x < -40 ? 'left' : null
	);
	let overlayOpacity = $derived(Math.min(Math.abs($pos.x) / SWIPE_THRESHOLD, 1));

	function onPointerDown(e: PointerEvent) {
		if (!isTop) return;
		// Ignore if triggered from a button that should handle its own click
		// (e.g. like/pass/view-profile buttons) but allow drags from those areas.
		dragging = true;
		didDrag = false;
		startX = e.clientX - $pos.x;
		startY = e.clientY - $pos.y;
		// Pointer capture ensures we receive move/up even when the pointer leaves
		// the card element (critical on iOS/Capacitor).
		try { cardEl.setPointerCapture(e.pointerId); } catch { /* ignore */ }
	}

	function onPointerMove(e: PointerEvent) {
		if (!dragging) return;
		const dx = e.clientX - startX;
		const dy = e.clientY - startY;
		if (Math.abs(dx) > TAP_THRESHOLD || Math.abs(dy) > TAP_THRESHOLD) {
			didDrag = true;
		}
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
		if (committed) return;
		committed = true;
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
		{fullScreen
			? 'inset: 0;'
			: 'left: 16px; right: 16px; top: min(114px, 14dvh); bottom: calc(var(--bottom-nav-clearance, 101px) + 16px);'}
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

		<!-- Transparent tap zone over the photo area to open the full profile.
		     pointer-events: none while dragging so the tap zone never swallows drag events. -->
		{#if isTop}
			<div
				class="absolute left-0 right-0 top-0"
				style="bottom: calc(max(env(safe-area-inset-bottom), 8px) + 160px); background: transparent; pointer-events: {dragging ? 'none' : 'auto'}; cursor: pointer;"
				role="button"
				tabindex="-1"
				aria-label="View full profile"
				onclick={(e) => { e.stopPropagation(); if (!didDrag) onviewprofile?.(profile.id); }}
			></div>
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
			style="bottom: calc(max(env(safe-area-inset-bottom), 8px) + {fullScreen ? 'var(--bottom-nav-clearance, 80px)' : '8px'}); gap: 12px;"
		>
			<!-- Name + age -->
			<div class="flex items-baseline gap-0">
				<span class="text-[26px] font-black leading-tight text-white tracking-tight">{profile.name}</span>
				<span class="text-[26px] font-black leading-tight text-white" style="letter-spacing: -1px;">, {profile.age}</span>
			</div>

			<!-- Tag pills + info rows in one compact block -->
			<div class="flex flex-col gap-8px" style="gap: 8px;">
				<!-- Tag pills -->
				{#if profile.tags.length > 0}
					<div class="flex flex-wrap gap-1.5">
						{#each profile.tags as tag}
							<span class="card-pill">{tag}</span>
						{/each}
					</div>
				{/if}

				<!-- Info rows -->
				<div class="flex flex-col gap-1">
					{#if profile.location}
						<div class="flex items-center gap-1.5">
							<i class="fi fi-rr-marker text-white" style="font-size: 13px; line-height: 1; flex-shrink: 0;"></i>
							<span class="text-[12px] font-medium text-white opacity-90">{profile.location}</span>
						</div>
					{/if}
					{#if profile.school}
						<div class="flex items-center gap-1.5">
							<i class="fi fi-rr-bank text-white" style="font-size: 13px; line-height: 1; flex-shrink: 0;"></i>
							<span class="text-[12px] font-medium text-white opacity-90">{profile.school}</span>
						</div>
					{/if}
					{#if profile.goals}
						<div class="flex items-center gap-1.5">
							<i class="fi fi-rr-star text-white" style="font-size: 13px; line-height: 1; flex-shrink: 0;"></i>
							<span class="text-[12px] font-medium text-white opacity-90">{profile.goals}</span>
						</div>
					{/if}
				</div>
			</div>

			<!-- Action buttons -->
			<div class="flex items-center gap-2" style="margin-top: 4px;">
				<!-- Pass -->
				<button
					class="flex flex-1 items-center justify-center gap-1.5 rounded-[50px]"
					style="height: 44px; background: rgba(255,255,255,0.92); backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);"
					onclick={(e) => { e.stopPropagation(); swipeLeft(); }}
				>
					<i class="fi fi-rr-cross" style="font-size: 15px; line-height: 1; color: #171717;"></i>
					<span class="text-[13px] font-semibold" style="color: #171717;">{$t('swipe.pass')}</span>
				</button>

				<!-- View Profile -->
				<button
					class="flex items-center justify-center rounded-[50px] px-4"
					style="height: 44px; background: rgba(255,255,255,0.92); backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);"
					onclick={(e) => { e.stopPropagation(); onviewprofile?.(profile.id); }}
				>
					<span class="text-[13px] font-semibold" style="color: #171717;">{$t('swipe.view_profile')}</span>
				</button>

				<!-- Like -->
				<button
					class="flex flex-1 items-center justify-center gap-1.5 rounded-[50px]"
					style="height: 44px; background: rgba(255,255,255,0.92); backdrop-filter: blur(8px); -webkit-backdrop-filter: blur(8px);"
					onclick={(e) => { e.stopPropagation(); swipeRight(); }}
				>
					<i class="fi fi-rr-heart" style="font-size: 15px; line-height: 1; color: #171717;"></i>
					<span class="text-[13px] font-semibold" style="color: #171717;">{$t('swipe.like')}</span>
				</button>
			</div>
		</div>
	</div>
</div>
