<template>
  <div v-if="!isHide" class="embed">
    <div class="embed__card">
      <div class="embed__card_top">
        <a target="_blank" href="https://x.com/stan_duel_duck">Stan Horuna</a>
        <ApproveSVG />
        <span>
          <a target="_blank" href="https://x.com/stan_duel_duck">@stan</a>
        </span>
        <DotSVG />
        <span>May 29</span>
      </div>

      <div class="embed__card_content">
        <p>
          85% of people still think
          <a target="_blank" href="https://twitter.com/hashtag/PredictionMarkets?src=hash&amp;ref_src=twsrc%5Etfw">#PredictionMarkets</a>
          are just casinos.
        </p>
        <p>
          But there’s no casino here. You play against each other. No bookmaker. No third party taking a cut.
        </p>
        <p>
          <a target="_blank" href="https://twitter.com/duel_duck?ref_src=twsrc%5Etfw">@Duel_Duck</a>
          is about fair duels, not lucky bets.
        </p>
        <a target="_blank" href="https://x.com/stan_duel_duck/status/1978495298595889274">
          <img src="https://pbs.twimg.com/media/G3UGrnDXEAAEwT0?format=png&name=small" alt="">
        </a>
      </div>

      <div class="embed__card_bottom">

      </div>
    </div>

    <div class="embed__buttons">
      <a target="_blank" href="https://arena.colosseum.org/projects/explore/duels">
        <span>Hackaton</span>
      </a>
      <a target="_blank" href="https://x.com/stan_duel_duck/status/1978495298595889274">
        <span>Share on</span>
        <TwitterSVG />
      </a>
    </div>

    <div class="embed__close" @click="closeHandler">
      <HideSVG />
    </div>
  </div>
</template>

<script setup>
import TwitterSVG from "./SVG/TwitterSVG.vue";
import ApproveSVG from "./SVG/ApproveSVG.vue";
import DotSVG from "./SVG/DotSVG.vue";
import HideSVG from "./SVG/HideSVG.vue";
import {onMounted, ref} from "vue";

const isHide = ref(false);

const closeHandler = () => {
  localStorage.setItem('twitterHide', 'true');
}

onMounted(() => {
  const data = localStorage.getItem('twitterHide');
  if (data) {
    isHide.value = true;
  }
})
</script>

<style scoped lang="scss">
.embed {
  position: fixed;
  bottom: 12px;
  left: 12px;
  &__buttons {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 8px;
    a {
      border-radius: 12px;
      background: #111;
      display: flex;
      height: 48px;
      padding: 0 16px;
      justify-content: center;
      align-items: center;
      gap: 8px;
      color: #FFF;
      font-size: 16px;
      font-weight: 500;
      line-height: 100%;
      outline: none;
      border: none;
      width: 100%;
      cursor: pointer;
      &:last-child {
        background: #000;
      }
    }
  }
  &__card {
    border-radius: 18px;
    border: 1px solid #111;
    background: #000;
    display: flex;
    width: 350px;
    padding: 12px;
    flex-direction: column;
    &_top {
      display: flex;
      align-items: center;
      gap: 3px;
      font-size: 14px;
      line-height: 100%;
      margin-bottom: 12px;
      a {
        color: #fff;
      }
      span {
        color: #7C7C7C;
        a {
          color: #7C7C7C;
        }
      }
    }
    &_content {
      font-size: 12px;
      a {
        color: #4D7CFF;
      }
      p {
        margin-bottom: 10px;
      }
      img {
        border: 1px solid #111;
        width: 100%;
        height: 180px;
        object-fit: cover;
        border-radius: 12px;
      }
    }
  }
  &__close {
    position: absolute;
    top: 8px;
    right: 8px;
    cursor: pointer;
  }
}

@media (max-width: 1200px) {
  .embed {
    display: none;
  }
}
</style>
