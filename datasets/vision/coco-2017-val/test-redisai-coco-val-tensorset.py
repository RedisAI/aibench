import csv

import redisai as rai
from skimage import io

import cv2
import numpy as np
from os import listdir
from os.path import isfile, join
import tqdm
import argparse
from pathlib import Path
from redisai.command_builder import Builder

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
    parser.add_argument('--input-val_dir', default='cropped-val2017')
    parser.add_argument('--host', default='localhost')
    parser.add_argument('--port', default=6379)
    args = parser.parse_args()
    print("Reading cropped scaled images to {}".format(args.input_val_dir))
    con = rai.Client(host=args.host, port=args.port)
    builder = Builder()
    filenames = [f for f in listdir(args.input_val_dir) if isfile(
        join(args.input_val_dir, f))]
    filenames.sort()
    with open("test.csv","w") as csvfile:
        csvwriter = csv.writer(csvfile)
    # spamwriter.writerow(['Spam'] * 5 + ['Baked Beans'])
    # spamwriter.writerow(['Spam', 'Lovely Spam', 'Wonderful Spam'])

        for filename in tqdm.tqdm(filenames):
            img_path = '{}/{}'.format(args.input_val_dir, filename)
            image = io.imread(img_path)
            tensorset_args = builder.tensorset(filename, image )
            final_args = ['WRITE','W1']
            tensorset_args.pop()
            tensorset_args.pop()
            final_args.extend(tensorset_args)
            csvwriter.writerow(final_args)
            # tensorset_args = [str(x) for x in tensorset_args ]
            #
            #
            #
            # # dag_args = ["AI.DAGRUN"]
            # csvfile.writelines(( "WRITE,TENSORSET," + ",".join(tensorset_args)+"\n"))
        # con.tensorset(filename, image)





