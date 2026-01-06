import type { FC } from 'hono/jsx'
import { Button } from './Button'
import { MemoGroup } from './MarkdownEditor'
import {
  IconChevronLeft,
  IconChevronRight,
  IconImage,
  IconTrash,
  IconArrowUpFromBracket,
  IconDownload,
  IconThumbtack,
  IconGrid,
  IconClose,
} from './icons'
import type { PageContext, CharacterImage, ImageGalleryState } from '../lib/types'
import { api } from '../lib/types'

type ProfileProps = {
  pc: PageContext
  gallery?: ImageGalleryState
}

// Gallery Modal Component (always renders container for OOB updates)
export const GalleryModal: FC<{ pc: PageContext; gallery: ImageGalleryState; oob?: boolean }> = ({ pc, gallery, oob }) => {
  const hasImages = gallery.images.length > 0
  
  return (
    <div
      id="gallery-modal"
      class="fixed inset-0 z-50 hidden items-center justify-center bg-black/60 backdrop-blur-sm"
      onclick="if(event.target === this) closeGalleryModal()"
      {...(oob ? { 'hx-swap-oob': 'true' } : {})}
    >
      {hasImages && (
        <div class="relative bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[80vh] overflow-hidden flex flex-col">
          {/* Header */}
          <div class="flex items-center justify-between p-4 border-b border-slate-200">
            <h3 class="text-lg font-semibold text-slate-700">画像一覧</h3>
            <button
              type="button"
              class="p-2 rounded-md text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
              onclick="closeGalleryModal()"
              title="閉じる"
            >
              <IconClose class="w-5 h-5" />
            </button>
          </div>
          {/* Image Grid */}
          <div class="p-4 overflow-y-auto">
            <div class="grid grid-cols-3 gap-2">
              {gallery.images.map((image, index) => (
                <button
                  type="button"
                  class={`relative aspect-square rounded-lg overflow-hidden border-2 transition-all hover:opacity-80 ${
                    index === gallery.currentIndex
                      ? 'border-blue-500 ring-2 ring-blue-200'
                      : 'border-transparent hover:border-slate-300'
                  }`}
                  hx-post={api(pc, `/api/image/select/${index}`)}
                  hx-swap="none"
                  hx-on--after-request="closeGalleryModal()"
                  title={image.filename}
                >
                  <img
                    src={image.data}
                    alt={image.filename}
                    class="w-full h-full object-cover"
                  />
                  {image.pinned && (
                    <div class="absolute top-1 right-1 p-1 bg-orange-500 text-white rounded-full">
                      <IconThumbtack class="w-3 h-3" />
                    </div>
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// Image Gallery Component (for OOB swaps)
export const ImageGallery: FC<{ pc: PageContext; gallery: ImageGalleryState; oob?: boolean }> = ({ pc, gallery, oob }) => {
  const hasImages = gallery.images.length > 0
  const currentImage = hasImages ? gallery.images[gallery.currentIndex] : null
  const canNavigate = gallery.images.length > 1
  
  return (
    <div id="image-gallery" class="flex flex-col gap-2" {...(oob ? { 'hx-swap-oob': 'true' } : {})}>
      <div class="aspect-square w-full overflow-hidden flex items-center justify-center rounded-lg bg-white shadow">
        {currentImage ? (
          <img
            src={currentImage.data}
            alt={currentImage.filename}
            class="w-full h-full object-cover"
          />
        ) : (
          <div class="flex items-center justify-center w-full h-full bg-slate-100 text-slate-300">
            <IconImage class="w-24 h-24" />
          </div>
        )}
      </div>
      {/* Image Controls */}
      <div class="flex gap-2 bg-white rounded-lg p-2 shadow">
        <Button
          variant="solid-plain"
          icon
          disabled={!canNavigate}
          title="前の画像"
          hxPost={canNavigate ? api(pc, '/api/image/prev') : undefined}
          hxSwap="none"
        >
          <IconChevronLeft />
        </Button>
        <Button
          variant="ghost"
          class="grow"
          disabled={!hasImages}
          onclick={hasImages ? 'openGalleryModal()' : undefined}
        >
          {hasImages ? `${gallery.currentIndex + 1} / ${gallery.images.length}` : '画像一覧'}
        </Button>
        {pc.isOwner && (
          <>
            <Button
              variant="ghost-red"
              icon
              title="画像を削除"
              disabled={!hasImages}
              hxPost={hasImages ? api(pc, '/api/image/delete') : undefined}
              hxSwap="none"
              hxConfirm="この画像を削除しますか？"
            >
              <IconTrash />
            </Button>
            <label class="relative">
              <input
                type="file"
                name="image"
                accept="image/*"
                class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                hx-post={api(pc, '/api/image/upload')}
                hx-trigger="change"
                hx-encoding="multipart/form-data"
                hx-swap="none"
              />
              <Button variant="ghost-blue" icon title="画像をアップロード" as="span">
                <IconArrowUpFromBracket />
              </Button>
            </label>
          </>
        )}
        {hasImages && currentImage && (
          <a
            href={currentImage.data}
            download={currentImage.filename}
            class="inline-flex items-center justify-center w-8 h-8 rounded-md text-green-600 hover:bg-green-50 transition-colors duration-150"
            title="画像をダウンロード"
          >
            <IconDownload />
          </a>
        )}
        {!hasImages && (
          <Button variant="ghost-green" icon disabled title="画像をダウンロード">
            <IconDownload />
          </Button>
        )}
        {pc.isOwner && (
          <Button
            variant={currentImage?.pinned ? 'solid-orange' : 'ghost-orange'}
            icon
            disabled={!hasImages}
            title="画像をピン留め"
            hxPost={hasImages ? api(pc, '/api/image/pin') : undefined}
            hxSwap="none"
          >
            <IconThumbtack />
          </Button>
        )}
        <Button
          variant="solid-plain"
          icon
          disabled={!canNavigate}
          title="次の画像"
          hxPost={canNavigate ? api(pc, '/api/image/next') : undefined}
          hxSwap="none"
        >
          <IconChevronRight />
        </Button>
      </div>
    </div>
  )
}

export const Profile: FC<ProfileProps> = ({ pc, gallery }) => {
  const readonly = !pc.isOwner || pc.preview
  const effectiveGallery = gallery || { images: [], currentIndex: 0 }

  return (
    <div class="flex flex-col gap-4">
      <ImageGallery pc={pc} gallery={effectiveGallery} />

      {/* Name Section */}
      <section class="bg-white rounded-lg shadow overflow-hidden" aria-label="character-profile">
        <div class="flex flex-col p-3">
          <input
            type="text"
            id="profile-name"
            name="profile_name"
            class="w-full text-4xl font-semibold text-blue-600 bg-transparent border-none outline-none hover:bg-slate-100 rounded-md transition-colors duration-150"
            placeholder="名前"
            readonly={readonly}
            hx-post={api(pc, '/api/profile/name/set')}
            hx-trigger="input changed delay:1500ms"
            hx-swap="none"
          />
          <input
            type="text"
            id="profile-ruby"
            name="profile_ruby"
            class="w-full text-lg text-slate-500 bg-transparent border-none outline-none hover:bg-slate-100 rounded-md transition-colors duration-150"
            placeholder="よみがな"
            readonly={readonly}
            hx-post={api(pc, '/api/profile/ruby/set')}
            hx-trigger="input changed delay:1500ms"
            hx-swap="none"
          />
        </div>
      </section>
      
      {/* Character-level memos */}
      <MemoGroup pc={pc} />
    </div>
  )
}
