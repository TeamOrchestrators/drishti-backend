from pathlib import Path
import numpy as np
from PIL import Image, ImageDraw

root = Path(__file__).parent
source = Image.open(root / 'drishti-methodology.png').convert('RGB')
rgb = np.asarray(source).astype(float)
w, h = source.size
# Keep the original panel fills and contents, including the two dashed containers.
scale = 4
mask = Image.new('L', (w * scale, h * scale))
draw = ImageDraw.Draw(mask)
panels = [(24,166,255,582,10), (273,166,511,582,10),
          (529,166,752,582,10), (771,163,1259,583,15),
          (1277,166,1515,582,10), (28,677,949,944,13),
          (1095,677,1509,944,12)]
for x0,y0,x1,y1,radius in panels:
    draw.rounded_rectangle((x0*scale,y0*scale,x1*scale,y1*scale),
                           radius=radius*scale, fill=255)
panel = np.asarray(mask.resize((w,h), Image.Resampling.LANCZOS)).astype(float)/255
# Exterior artwork is dark navy on an almost-white canvas. Recover coverage
# and remove the white matte only from exterior antialiased edges.
minimum = rgb.min(axis=2)
alpha = np.clip((255-minimum)/250, 0, 1)
alpha[minimum > 235] = 0
alpha[minimum < 35] = 1
unmatted = np.clip((rgb-255*(1-alpha[:,:,None])) /
                    np.maximum(alpha[:,:,None], 1e-8), 0, 255)
final_alpha = panel + (1-panel)*alpha
premultiplied = rgb*panel[:,:,None] + unmatted*(alpha*(1-panel))[:,:,None]
final_rgb = np.clip(premultiplied / np.maximum(final_alpha[:,:,None],1e-8),0,255)
rgba = np.dstack((final_rgb,final_alpha*255)).round().astype('uint8')
result = Image.fromarray(rgba)
result.save(root / 'drishti-methodology-transparent.png')
# QA preview on a dark neutral canvas reveals white fringes and missing details.
preview = Image.new('RGBA', source.size, '#667080')
preview.alpha_composite(result)
preview.convert('RGB').save(root / 'drishti-methodology-transparency-check.png')
inside = panel == 1
assert np.array_equal(rgba[:,:,:3][inside], np.asarray(source)[inside])
assert rgba[0,0,3] == 0
assert rgba[200,100,3] == 255
print(f'Output: {w}x{h} RGBA; fully transparent pixels: {(rgba[:,:,3]==0).sum()}; panel interiors unchanged.')
