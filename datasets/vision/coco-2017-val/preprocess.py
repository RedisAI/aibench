import cv2
import numpy as np
from os import listdir
from os.path import isfile, join
import argparse
from pathlib import Path
from tqdm import tqdm

"""
 script to pre process the specified input images. For each image:
    (1) Resize the image so its smaller side is 256 pixels long
    (2) Take a random 224 x 224 crop to the scaled image
"""

def image_min_resize(image, smaller_size, inter=cv2.INTER_AREA):
    # initialize the dimensions of the image to be resized and
    # grab the image size
    dim = None
    (h, w) = image.shape[:2]

    minv = h if h < w else w
    if minv == h:
        r = smaller_size / float(h)
        dim = (int(w * r), smaller_size)
    else:
        r = smaller_size / float(w)
        dim = (smaller_size, int(h * r))

    # resize the image
    resized = cv2.resize(image, dim, interpolation=inter)

    # return the resized image
    return resized


def get_random_crop(image, crop_height, crop_width):

    max_x = image.shape[1] - crop_width
    max_y = image.shape[0] - crop_height

    x = np.random.randint(0, max_x)
    y = np.random.randint(0, max_y)

    crop = image[y: y + crop_height, x: x + crop_width]

    return crop


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Script to pre process the specified input images so that they are ready for ResNet-50 model")
    parser.add_argument('--input-val_dir', default='val2017')
    parser.add_argument('--output-val_dir', default='cropped-val2017')
    parser.add_argument('--random-seed', type=int, default=12345)
    parser.add_argument('--re-use-factor', type=int, default=10)
    args = parser.parse_args()
    print("Using random seed {} to take a random 224 x 224 crop to the scaled image".format(args.random_seed))
    print("Saving cropped scaled images to {}".format(args.output_val_dir))

    np.random.seed(args.random_seed)

    # ensure output val path exists
    Path(args.output_val_dir).mkdir(parents=True, exist_ok=True)

    filenames = [f for f in listdir(args.input_val_dir) if isfile(
        join(args.input_val_dir, f))]
    filenames.sort()
    total_images = len(filenames) * args.re_use_factor
    print("Total images to save {}".format(total_images))
    progress = tqdm(unit="images", total=total_images)

    for filename in filenames:
        img = cv2.imread('{}/{}'.format(args.input_val_dir, filename))
        resized_img = image_min_resize(img, 256)
        for sequence in range(1,args.re_use_factor+1):
            cropped_img = get_random_crop(resized_img, 224, 224)
            new_filename = "rep{}_{}".format(sequence,filename)
            cv2.imwrite('{}/{}'.format(args.output_val_dir, new_filename), cropped_img)
            progress.update()
    progress.close()
