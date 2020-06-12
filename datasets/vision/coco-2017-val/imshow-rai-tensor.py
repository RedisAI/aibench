from skimage import io
import redisai as rai

con = rai.Client()

image2 = con.tensorget('000000019042.jpg')
print(image2.shape)
io.imshow(image2)
io.show()