---
title: Bakery
nav_order: 1
has_children: true
has_toc: false
---

# Bakery

Bakery is a manifest manipulation service for HLS and DASH


## What's new

Follow the updates to our API here!

<ul>
  {% for post in site.posts limit:15 %}
	{% if post.category contains "whats-new" %}
    <li>
      <span class="post-date">{{ post.date | date: "%B %d, %Y" }}</span> <a href="{{ site.baseurl }}{{ post.url }}">{{ post.title }}</a>
    </li>
    {% endif %}
  {% endfor %}
</ul>

## Getting Started

Unsure how to get started? Check out our quick start tutorial!
<ul>
  {% for post in site.posts limit:15 %}
	{% if post.category contains "quick-start" %}
    <li>
      <span class="post-date">{{ post.date | date: "%B %d, %Y" }}</span> <a href="{{ site.baseurl }}{{ post.url }}">{{ post.title }}</a>
    </li>
    {% endif %}
  {% endfor %}
</ul>

## Filters

Check out our Filters <a href="/filters">here</a>!


## Help

You can find the source code for Bakery at GitHub:
[bakery][bakery]

[bakery]: https://github.com/cbsinteractive/bakery

If you have any questions regarding Bakery, please reach out in the [#i-vidtech-mediahub](slack://channel?team={cbs}&id={i-vidtech-mediahub}) channel.